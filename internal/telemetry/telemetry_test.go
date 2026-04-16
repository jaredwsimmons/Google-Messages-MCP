package telemetry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDisabledClientDoesNothing(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
	}))
	defer srv.Close()

	c := New(t.TempDir(), "test", false)
	c.SetEndpoint(srv.URL)
	c.MaybeSend(context.Background(), PlatformStatus{GoogleMessages: true})

	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("disabled client made %d HTTP calls; want 0", got)
	}
}

func TestEnabledClientSendsOnce(t *testing.T) {
	t.Parallel()

	var calls int32
	var captured payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
	}))
	defer srv.Close()

	c := New(t.TempDir(), "v0.2.0", true)
	c.SetEndpoint(srv.URL)
	c.MaybeSend(context.Background(), PlatformStatus{GoogleMessages: true, WhatsApp: true})

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
	if captured.Version != "v0.2.0" {
		t.Fatalf("version = %q; want v0.2.0", captured.Version)
	}
	if !captured.Platforms.GoogleMessages || !captured.Platforms.WhatsApp || captured.Platforms.Signal {
		t.Fatalf("platforms not propagated correctly: %+v", captured.Platforms)
	}
	if len(captured.InstallID) != 32 {
		t.Fatalf("install id wrong length: %q", captured.InstallID)
	}
}

func TestRespectsHeartbeatPeriod(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
	}))
	defer srv.Close()

	c := New(t.TempDir(), "test", true)
	c.SetEndpoint(srv.URL)

	for i := 0; i < 5; i++ {
		c.MaybeSend(context.Background(), PlatformStatus{})
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 call within period, got %d", got)
	}
}

func TestInstallIDIsStable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := New(dir, "test", true)

	id1, err := c.installID()
	if err != nil {
		t.Fatalf("first install id: %v", err)
	}
	id2, err := c.installID()
	if err != nil {
		t.Fatalf("second install id: %v", err)
	}
	if id1 != id2 {
		t.Fatalf("install id changed: %q vs %q", id1, id2)
	}
	if len(id1) != 32 {
		t.Fatalf("install id length = %d; want 32", len(id1))
	}
}

func TestNetworkErrorIsSilent(t *testing.T) {
	t.Parallel()

	c := New(t.TempDir(), "test", true)
	c.SetEndpoint("http://127.0.0.1:1") // closed port
	c.httpc = &http.Client{Timeout: 50 * time.Millisecond}

	// Should not panic or block.
	done := make(chan struct{})
	go func() {
		c.MaybeSend(context.Background(), PlatformStatus{})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("MaybeSend blocked too long on network failure")
	}
}
