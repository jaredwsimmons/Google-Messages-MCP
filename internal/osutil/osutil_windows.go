//go:build windows

package osutil

import (
	"errors"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// EnforcesFileModes is false on Windows. Go derives a file's permission bits
// from its read-only attribute alone (0666 or 0444), so a check that a file
// is 0600 can never pass; the data directory's ACLs are what protect it.
const EnforcesFileModes = false

// SetUmask does nothing on Windows, which has no process umask. New files
// inherit the ACLs of the directory they are created in.
func SetUmask(int) (restore func()) {
	return func() {}
}

// IsConnRefused reports whether err means nothing was listening on the
// address. Winsock reports a refused connection as WSAECONNREFUSED (10061),
// which Go does not map to syscall.ECONNREFUSED, so both are checked.
func IsConnRefused(err error) bool {
	return errors.Is(err, windows.WSAECONNREFUSED) || errors.Is(err, syscall.ECONNREFUSED)
}

// lockOffsetHigh places the one-byte lock far past any real file content.
// Windows byte-range locks are mandatory, so locking the bytes a lock file's
// diagnostic record lives in would stop other processes from reading it.
// Locking beyond the end of a file is allowed.
const lockOffsetHigh = 0x7fffffff

// LockExclusive takes a non-blocking exclusive lock on file. It returns
// ErrLocked when another handle holds the lock.
func LockExclusive(file *os.File) error {
	overlapped := windows.Overlapped{OffsetHigh: lockOffsetHigh}
	err := windows.LockFileEx(windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0, 1, 0, &overlapped)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return ErrLocked
	}
	return err
}

// Unlock releases a lock taken by LockExclusive.
func Unlock(file *os.File) error {
	overlapped := windows.Overlapped{OffsetHigh: lockOffsetHigh}
	return windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, &overlapped)
}

// SyncDir only confirms that path exists on Windows. FlushFileBuffers needs
// a handle with write access, which os.Open cannot give a directory, so
// File.Sync on a directory fails with "Access is denied". Windows offers no
// directory fsync for this code to call instead.
func SyncDir(path string) error {
	_, err := os.Stat(path)
	return err
}

// SyncFile flushes an existing file's contents to disk. It opens the file
// for writing because FlushFileBuffers needs a handle with write access;
// File.Sync on a read-only handle fails with "Access is denied".
func SyncFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	syncErr := file.Sync()
	closeErr := file.Close()
	return errors.Join(syncErr, closeErr)
}

// AvailableBytes returns the free space the caller can use (after any disk
// quota) on the volume that holds path.
func AvailableBytes(path string) (uint64, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var available, total, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(pathPtr, &available, &total, &totalFree); err != nil {
		return 0, err
	}
	return available, nil
}

// fileRetryWindow bounds how long Rename and ReadFile retry. It matches the
// two seconds Go's own toolchain allows for the same Windows behavior
// (cmd/go/internal/robustio).
const fileRetryWindow = 2 * time.Second

// isTransientFileError reports the errors Windows returns while another
// handle has a file open without FILE_SHARE_DELETE, as when one process reads
// a file that another is replacing with a rename.
func isTransientFileError(err error) bool {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}
	return errno == windows.ERROR_ACCESS_DENIED || errno == windows.ERROR_SHARING_VIOLATION
}

func retryTransient(op func() error) error {
	deadline := time.Now().Add(fileRetryWindow)
	delay := time.Millisecond
	for {
		err := op()
		if err == nil || !isTransientFileError(err) || time.Now().After(deadline) {
			return err
		}
		time.Sleep(delay)
		if delay < 64*time.Millisecond {
			delay *= 2
		}
	}
}

// Rename is os.Rename, retried for up to two seconds while another handle
// holds either file open. Replacing a file another process is reading fails
// on Windows with "Access is denied" until that reader closes it.
func Rename(oldpath, newpath string) error {
	return retryTransient(func() error { return os.Rename(oldpath, newpath) })
}

// ReadFile is os.ReadFile, retried for up to two seconds while another
// process is replacing the file, which Windows reports as a sharing
// violation.
func ReadFile(name string) ([]byte, error) {
	var data []byte
	err := retryTransient(func() error {
		var readErr error
		data, readErr = os.ReadFile(name)
		return readErr
	})
	return data, err
}
