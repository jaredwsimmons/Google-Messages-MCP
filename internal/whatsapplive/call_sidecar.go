package whatsapplive

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.mau.fi/whatsmeow"
)

const (
	whatsAppCallSidecarEnv          = "OPENMESSAGES_WHATSAPP_CALL_SIDECAR"
	whatsAppCallSidecarTraceFileEnv = "OPENMESSAGES_WHATSAPP_CALL_TRACE_FILE"
	whatsAppCallSidecarRejectEnv    = "OPENMESSAGES_WHATSAPP_CALL_REJECT_OFFERS"
	callSidecarShutdownWaitTimeout  = 2 * time.Second
)

var startCallJSONSidecar = func(cli *whatsmeow.Client, cmd *exec.Cmd) (callSidecarProcess, error) {
	return cli.StartCallJSONSidecar(cmd)
}

type CallSidecarSnapshot struct {
	Configured bool     `json:"configured"`
	Running    bool     `json:"running"`
	Command    []string `json:"command,omitempty"`
	LastError  string   `json:"last_error,omitempty"`
}

type callSidecarConfig struct {
	Command      []string
	TraceFile    string
	RejectOffers bool
}

type callSidecarProcess interface {
	HandleSignaling(context.Context, *whatsmeow.CallSignalingEvent)
	Serve(context.Context) error
	Stderr() io.Reader
	Close() error
	Wait() error
}

type callSidecarManager struct {
	logger     zerolog.Logger
	configured bool
	config     callSidecarConfig

	mu          sync.RWMutex
	client      *whatsmeow.Client
	process     callSidecarProcess
	serveCancel context.CancelFunc
	serveDone   chan struct{}
	waitDone    chan struct{}
	lastError   string
}

func newCallSidecarManager(logger zerolog.Logger) *callSidecarManager {
	cfg, configured, err := loadCallSidecarConfigFromEnv()
	if !configured && err == nil {
		return nil
	}
	manager := &callSidecarManager{
		logger:     logger,
		configured: configured,
		config:     cfg,
	}
	if err != nil {
		manager.lastError = err.Error()
	}
	return manager
}

func loadCallSidecarConfigFromEnv() (callSidecarConfig, bool, error) {
	raw := strings.TrimSpace(os.Getenv(whatsAppCallSidecarEnv))
	if raw == "" {
		return callSidecarConfig{}, false, nil
	}
	command, err := parseCallSidecarCommand(raw)
	if err != nil {
		return callSidecarConfig{}, true, err
	}
	return callSidecarConfig{
		Command:      command,
		TraceFile:    strings.TrimSpace(os.Getenv(whatsAppCallSidecarTraceFileEnv)),
		RejectOffers: envTruthy(whatsAppCallSidecarRejectEnv),
	}, true, nil
}

func parseCallSidecarCommand(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("whatsapp call sidecar command is empty")
	}
	if strings.HasPrefix(raw, "[") {
		var command []string
		if err := json.Unmarshal([]byte(raw), &command); err != nil {
			return nil, err
		}
		if len(command) == 0 {
			return nil, errors.New("whatsapp call sidecar command is empty")
		}
		return command, nil
	}
	command := strings.Fields(raw)
	if len(command) == 0 {
		return nil, errors.New("whatsapp call sidecar command is empty")
	}
	return command, nil
}

func envTruthy(name string) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func (cfg callSidecarConfig) commandArgs() []string {
	if len(cfg.Command) == 0 {
		return nil
	}
	args := append([]string(nil), cfg.Command...)
	if cfg.TraceFile != "" {
		args = append(args, "--trace-file", cfg.TraceFile)
	}
	if cfg.RejectOffers {
		args = append(args, "--reject-offers")
	}
	return args
}

func (manager *callSidecarManager) attach(cli *whatsmeow.Client) {
	if manager == nil || cli == nil {
		return
	}

	manager.mu.RLock()
	configured := manager.configured
	command := manager.config.commandArgs()
	lastError := manager.lastError
	manager.mu.RUnlock()
	if !configured {
		return
	}
	if len(command) == 0 {
		if lastError != "" {
			manager.logger.Warn().Str("error", lastError).Msg("WhatsApp call sidecar is configured but invalid")
		}
		return
	}

	manager.stop()

	cmd := exec.Command(command[0], command[1:]...)
	sidecar, err := startCallJSONSidecar(cli, cmd)
	if err != nil {
		manager.recordError("start WhatsApp call sidecar", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	serveDone := make(chan struct{})
	waitDone := make(chan struct{})

	manager.mu.Lock()
	manager.client = cli
	manager.process = sidecar
	manager.serveCancel = cancel
	manager.serveDone = serveDone
	manager.waitDone = waitDone
	manager.lastError = ""
	manager.mu.Unlock()

	cli.CallSignalingHandler = sidecar.HandleSignaling

	if stderr := sidecar.Stderr(); stderr != nil {
		go manager.logStderr(stderr)
	}
	go func() {
		defer close(serveDone)
		if err := sidecar.Serve(ctx); err != nil && !errors.Is(err, context.Canceled) {
			manager.recordError("serve WhatsApp call sidecar", err)
		}
	}()
	go func() {
		defer close(waitDone)
		manager.handleWaitResult(cli, sidecar, sidecar.Wait())
	}()
}

func (manager *callSidecarManager) stop() {
	if manager == nil {
		return
	}

	manager.mu.Lock()
	client := manager.client
	process := manager.process
	cancel := manager.serveCancel
	serveDone := manager.serveDone
	waitDone := manager.waitDone
	manager.client = nil
	manager.process = nil
	manager.serveCancel = nil
	manager.serveDone = nil
	manager.waitDone = nil
	manager.mu.Unlock()

	if client != nil {
		client.CallSignalingHandler = nil
	}
	if cancel != nil {
		cancel()
	}
	if process != nil {
		if err := process.Close(); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			manager.recordError("close WhatsApp call sidecar", err)
		}
	}
	waitForCallSidecarShutdown(serveDone)
	waitForCallSidecarShutdown(waitDone)
}

func waitForCallSidecarShutdown(done <-chan struct{}) {
	if done == nil {
		return
	}
	select {
	case <-done:
	case <-time.After(callSidecarShutdownWaitTimeout):
	}
}

func (manager *callSidecarManager) snapshot() *CallSidecarSnapshot {
	if manager == nil {
		return nil
	}
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return &CallSidecarSnapshot{
		Configured: manager.configured,
		Running:    manager.process != nil,
		Command:    manager.config.commandArgs(),
		LastError:  manager.lastError,
	}
}

func (manager *callSidecarManager) recordError(action string, err error) {
	if manager == nil || err == nil {
		return
	}
	message := strings.TrimSpace(action)
	if message == "" {
		message = err.Error()
	} else {
		message += ": " + err.Error()
	}
	manager.mu.Lock()
	manager.lastError = message
	manager.mu.Unlock()
	manager.logger.Warn().Err(err).Msg(action)
}

func (manager *callSidecarManager) logStderr(stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		manager.logger.Info().Str("component", "whatsapp_call_sidecar").Msg(line)
	}
	if err := scanner.Err(); err != nil {
		manager.recordError("read WhatsApp call sidecar stderr", err)
	}
}

func (manager *callSidecarManager) handleWaitResult(cli *whatsmeow.Client, process callSidecarProcess, err error) {
	manager.mu.Lock()
	if manager.process == process {
		manager.process = nil
		manager.client = nil
		manager.serveCancel = nil
		manager.serveDone = nil
		manager.waitDone = nil
	}
	manager.mu.Unlock()

	if cli != nil {
		cli.CallSignalingHandler = nil
	}
	if err != nil {
		manager.recordError("wait for WhatsApp call sidecar", err)
	}
}
