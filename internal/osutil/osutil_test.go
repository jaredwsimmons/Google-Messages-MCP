package osutil

import (
	"bytes"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
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

func TestSQLiteFileURIForm(t *testing.T) {
	cases := map[string]string{
		"/data/messages.db": "file:///data/messages.db?mode=ro",
	}
	if runtime.GOOS == "windows" {
		cases = map[string]string{
			`C:\data\messages.db`:      "file:///C:/data/messages.db?mode=ro",
			`C:\Users\Jo Ann\store.db`: "file:///C:/Users/Jo%20Ann/store.db?mode=ro",
		}
	}
	for path, want := range cases {
		if got := SQLiteFileURI(path, "mode=ro"); got != want {
			t.Errorf("SQLiteFileURI(%q) = %q, want %q", path, got, want)
		}
	}
}

// Opening through the URI is what failed on Windows ("out of memory"), so
// this checks a real open rather than only the string.
func TestSQLiteFileURIOpensReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dir with space", "messages.db")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Exec(`CREATE TABLE t (v TEXT); INSERT INTO t VALUES ('hello')`); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	reader, err := sql.Open("sqlite", SQLiteFileURI(path, "mode=ro"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var got string
	if err := reader.QueryRow(`SELECT v FROM t`).Scan(&got); err != nil {
		t.Fatalf("read through %s: %v", SQLiteFileURI(path, "mode=ro"), err)
	}
	if got != "hello" {
		t.Fatalf("read %q, want hello", got)
	}
	if _, err := reader.Exec(`INSERT INTO t VALUES ('nope')`); err == nil {
		t.Fatal("write through a mode=ro URI succeeded")
	}
}

// Readers must never see a failed read or a partial file while another
// goroutine keeps replacing the file by rename, which on Windows needs the
// retries in Rename and ReadFile.
func TestRenameAndReadFileUnderConcurrentReplace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	versions := [][]byte{bytes.Repeat([]byte("a"), 4096), bytes.Repeat([]byte("b"), 4096)}
	if err := os.WriteFile(path, versions[0], 0o600); err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			tmp := filepath.Join(dir, "session.tmp")
			if err := os.WriteFile(tmp, versions[i%2], 0o600); err != nil {
				errs <- err
				return
			}
			if err := Rename(tmp, path); err != nil {
				errs <- err
				return
			}
		}
	}()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		data, err := ReadFile(path)
		if err != nil {
			close(stop)
			wg.Wait()
			t.Fatalf("ReadFile during replace: %v", err)
		}
		if !bytes.Equal(data, versions[0]) && !bytes.Equal(data, versions[1]) {
			close(stop)
			wg.Wait()
			t.Fatalf("ReadFile returned a partial file of %d bytes", len(data))
		}
	}
	close(stop)
	wg.Wait()
	select {
	case err := <-errs:
		t.Fatalf("Rename during reads: %v", err)
	default:
	}
}

// The v2 migration flushes its staged store this way. On Windows the flush
// needs a writable handle, so this also covers that SyncFile opens one.
func TestSyncFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.sqlite3")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := SyncFile(path); err != nil {
		t.Fatalf("SyncFile(existing file): %v", err)
	}
	if err := SyncFile(path + ".missing"); err == nil {
		t.Fatal("SyncFile(missing file) returned nil")
	}
}
