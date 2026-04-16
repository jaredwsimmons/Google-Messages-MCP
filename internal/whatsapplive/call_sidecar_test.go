package whatsapplive

import (
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"go.mau.fi/whatsmeow"

	"github.com/maxghenis/openmessage/internal/db"
)

type fakeCallSidecar struct {
	mu       sync.Mutex
	closeCh  chan struct{}
	stderr   io.Reader
	closed   bool
	closeCnt int
	waitCnt  int
	serveCnt int
}

func newFakeCallSidecar(stderr string) *fakeCallSidecar {
	return &fakeCallSidecar{
		closeCh: make(chan struct{}),
		stderr:  strings.NewReader(stderr),
	}
}

func (fake *fakeCallSidecar) HandleSignaling(context.Context, *whatsmeow.CallSignalingEvent) {}

func (fake *fakeCallSidecar) Serve(ctx context.Context) error {
	fake.mu.Lock()
	fake.serveCnt++
	fake.mu.Unlock()
	<-ctx.Done()
	return ctx.Err()
}

func (fake *fakeCallSidecar) Stderr() io.Reader {
	return fake.stderr
}

func (fake *fakeCallSidecar) Close() error {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.closeCnt++
	if !fake.closed {
		close(fake.closeCh)
		fake.closed = true
	}
	return nil
}

func (fake *fakeCallSidecar) Wait() error {
	fake.mu.Lock()
	fake.waitCnt++
	fake.mu.Unlock()
	<-fake.closeCh
	return nil
}

func (fake *fakeCallSidecar) counts() (closeCnt, waitCnt, serveCnt int) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	return fake.closeCnt, fake.waitCnt, fake.serveCnt
}

func newTestBridge(t *testing.T) *Bridge {
	t.Helper()
	store, err := db.New(filepath.Join(t.TempDir(), "messages.db"))
	if err != nil {
		t.Fatalf("db.New(): %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	bridge, err := New(filepath.Join(t.TempDir(), "whatsapp-session.db"), store, zerolog.Nop(), Callbacks{})
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	t.Cleanup(func() { _ = bridge.Close() })
	return bridge
}

func TestParseCallSidecarCommandJSON(t *testing.T) {
	got, err := parseCallSidecarCommand(`["/tmp/whatsmeow-call-sidecar","--reject-offers"]`)
	if err != nil {
		t.Fatalf("parseCallSidecarCommand(): %v", err)
	}
	want := []string{"/tmp/whatsmeow-call-sidecar", "--reject-offers"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}

func TestBridgeStartsConfiguredCallSidecar(t *testing.T) {
	traceFile := filepath.Join(t.TempDir(), "calls.ndjson")
	t.Setenv(whatsAppCallSidecarEnv, `["/tmp/whatsmeow-call-sidecar"]`)
	t.Setenv(whatsAppCallSidecarTraceFileEnv, traceFile)
	t.Setenv(whatsAppCallSidecarRejectEnv, "1")

	originalStart := startCallJSONSidecar
	fake := newFakeCallSidecar("ready\n")
	var capturedArgs []string
	startCallJSONSidecar = func(cli *whatsmeow.Client, cmd *exec.Cmd) (callSidecarProcess, error) {
		if cli == nil {
			t.Fatal("expected client")
		}
		capturedArgs = append([]string(nil), cmd.Args...)
		return fake, nil
	}
	t.Cleanup(func() { startCallJSONSidecar = originalStart })

	bridge := newTestBridge(t)

	wantArgs := []string{"/tmp/whatsmeow-call-sidecar", "--trace-file", traceFile, "--reject-offers"}
	if !reflect.DeepEqual(capturedArgs, wantArgs) {
		t.Fatalf("command args = %#v, want %#v", capturedArgs, wantArgs)
	}
	if bridge.client == nil || bridge.client.CallSignalingHandler == nil {
		t.Fatal("expected call signaling handler to be attached")
	}

	status := bridge.Status()
	if status.CallSidecar == nil {
		t.Fatal("expected call sidecar status")
	}
	if !status.CallSidecar.Configured || !status.CallSidecar.Running {
		t.Fatalf("call sidecar status = %+v, want configured and running", *status.CallSidecar)
	}

	if err := bridge.Close(); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	closeCnt, waitCnt, serveCnt := fake.counts()
	if closeCnt != 1 {
		t.Fatalf("close count = %d, want 1", closeCnt)
	}
	if waitCnt != 1 {
		t.Fatalf("wait count = %d, want 1", waitCnt)
	}
	if serveCnt != 1 {
		t.Fatalf("serve count = %d, want 1", serveCnt)
	}

	status = bridge.Status()
	if status.CallSidecar == nil || status.CallSidecar.Running {
		t.Fatalf("call sidecar status after close = %+v, want stopped", status.CallSidecar)
	}
}

func TestBridgeCallSidecarReportsConfigError(t *testing.T) {
	t.Setenv(whatsAppCallSidecarEnv, "[")

	bridge := newTestBridge(t)
	status := bridge.Status()
	if status.CallSidecar == nil {
		t.Fatal("expected call sidecar status")
	}
	if !status.CallSidecar.Configured {
		t.Fatalf("call sidecar status = %+v, want configured", *status.CallSidecar)
	}
	if status.CallSidecar.Running {
		t.Fatalf("call sidecar status = %+v, want not running", *status.CallSidecar)
	}
	if status.CallSidecar.LastError == "" || (!strings.Contains(strings.ToLower(status.CallSidecar.LastError), "json") && !strings.Contains(strings.ToLower(status.CallSidecar.LastError), "unexpected end")) {
		t.Fatalf("last error = %q, want parse error", status.CallSidecar.LastError)
	}
}

func TestBridgeResetClientLockedRestartsCallSidecar(t *testing.T) {
	t.Setenv(whatsAppCallSidecarEnv, `["/tmp/whatsmeow-call-sidecar"]`)

	originalStart := startCallJSONSidecar
	var (
		mu       sync.Mutex
		started  []*fakeCallSidecar
		commands [][]string
	)
	startCallJSONSidecar = func(_ *whatsmeow.Client, cmd *exec.Cmd) (callSidecarProcess, error) {
		mu.Lock()
		defer mu.Unlock()
		sidecar := newFakeCallSidecar("")
		started = append(started, sidecar)
		commands = append(commands, append([]string(nil), cmd.Args...))
		return sidecar, nil
	}
	t.Cleanup(func() { startCallJSONSidecar = originalStart })

	bridge := newTestBridge(t)

	bridge.mu.Lock()
	err := bridge.resetClientLocked()
	bridge.mu.Unlock()
	if err != nil {
		t.Fatalf("resetClientLocked(): %v", err)
	}

	mu.Lock()
	if len(started) != 2 {
		mu.Unlock()
		t.Fatalf("sidecar starts = %d, want 2", len(started))
	}
	first := started[0]
	second := started[1]
	if !reflect.DeepEqual(commands[0], []string{"/tmp/whatsmeow-call-sidecar"}) {
		mu.Unlock()
		t.Fatalf("first command = %#v, want binary only", commands[0])
	}
	if !reflect.DeepEqual(commands[1], []string{"/tmp/whatsmeow-call-sidecar"}) {
		mu.Unlock()
		t.Fatalf("second command = %#v, want binary only", commands[1])
	}
	mu.Unlock()

	waitForSidecarCounts(t, second, 0, 1, 1)

	firstCloseCnt, firstWaitCnt, _ := first.counts()
	if firstCloseCnt != 1 || firstWaitCnt != 1 {
		t.Fatalf("first sidecar counts = close:%d wait:%d, want 1/1", firstCloseCnt, firstWaitCnt)
	}
	secondCloseCnt, secondWaitCnt, secondServeCnt := second.counts()
	if secondCloseCnt != 0 || secondWaitCnt != 1 || secondServeCnt != 1 {
		t.Fatalf("second sidecar counts before final close = close:%d wait:%d serve:%d, want 0/1/1", secondCloseCnt, secondWaitCnt, secondServeCnt)
	}
}

func waitForSidecarCounts(t *testing.T, sidecar *fakeCallSidecar, closeCnt, waitCnt, serveCnt int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		gotClose, gotWait, gotServe := sidecar.counts()
		if gotClose == closeCnt && gotWait == waitCnt && gotServe == serveCnt {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	gotClose, gotWait, gotServe := sidecar.counts()
	t.Fatalf("sidecar counts = close:%d wait:%d serve:%d, want %d/%d/%d", gotClose, gotWait, gotServe, closeCnt, waitCnt, serveCnt)
}
