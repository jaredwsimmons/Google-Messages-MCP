package osutil

import (
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLockExclusiveRefusesSecondHolder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "instance.lock")
	first, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := os.OpenFile(path, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	if err := LockExclusive(first); err != nil {
		t.Fatalf("first lock: %v", err)
	}
	if err := LockExclusive(second); !errors.Is(err, ErrLocked) {
		t.Fatalf("second lock while first is held: got %v, want ErrLocked", err)
	}
	if err := Unlock(first); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	if err := LockExclusive(second); err != nil {
		t.Fatalf("second lock after unlock: %v", err)
	}
	if err := Unlock(second); err != nil {
		t.Fatalf("unlock second: %v", err)
	}
}

// The lock file carries a diagnostic record; holding the lock must not stop
// another process from reading it (Windows byte-range locks are mandatory).
func TestLockedFileContentStaysReadable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "instance.lock")
	holder, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Close()
	if err := LockExclusive(holder); err != nil {
		t.Fatal(err)
	}
	defer Unlock(holder)
	if _, err := holder.Write([]byte(`{"pid":1}`)); err != nil {
		t.Fatalf("write while holding lock: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read while another handle holds the lock: %v", err)
	}
	if string(got) != `{"pid":1}` {
		t.Fatalf("read %q", got)
	}
}

func TestIsConnRefused(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	_, err = client.Get("http://" + addr + "/api/status")
	if err == nil {
		t.Fatal("expected an error from a closed port")
	}
	if !IsConnRefused(err) {
		t.Fatalf("IsConnRefused(%v) = false, want true", err)
	}
	if IsConnRefused(errors.New("some other failure")) {
		t.Fatal("IsConnRefused matched an unrelated error")
	}
	if IsConnRefused(nil) {
		t.Fatal("IsConnRefused matched nil")
	}
}

func TestSyncDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := SyncDir(dir); err != nil {
		t.Fatalf("SyncDir(existing dir): %v", err)
	}
	if err := SyncDir(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("SyncDir(missing dir) returned nil")
	}
}

func TestAvailableBytes(t *testing.T) {
	available, err := AvailableBytes(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if available == 0 {
		t.Fatal("AvailableBytes reported 0 on the test volume")
	}
}

func TestSetUmaskRestores(t *testing.T) {
	restore := SetUmask(0o077)
	path := filepath.Join(t.TempDir(), "private")
	err := os.WriteFile(path, []byte("x"), 0o666)
	restore()
	if err != nil {
		t.Fatal(err)
	}
	if !EnforcesFileModes {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("file created under umask 077 has mode %04o, want 0600", got)
	}
}
