//go:build !windows

package osutil

import (
	"errors"
	"fmt"
	"math"
	"os"
	"syscall"
)

// EnforcesFileModes reports whether Unix permission bits on files are real
// access controls that callers can set and check.
const EnforcesFileModes = true

// SetUmask sets the process umask and returns a function that restores the
// previous one.
func SetUmask(mask int) (restore func()) {
	previous := syscall.Umask(mask)
	return func() { syscall.Umask(previous) }
}

// IsConnRefused reports whether err means nothing was listening on the
// address, so a request certainly did not reach a server.
func IsConnRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}

// LockExclusive takes a non-blocking exclusive advisory lock on file. It
// returns ErrLocked when another open file description holds the lock.
func LockExclusive(file *os.File) error {
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return ErrLocked
		}
		return err
	}
	return nil
}

// Unlock releases a lock taken by LockExclusive.
func Unlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

// SyncDir flushes a directory's entries to disk, making a preceding create,
// rename or remove inside it durable.
func SyncDir(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	return errors.Join(syncErr, closeErr)
}

// AvailableBytes returns the free space an unprivileged caller can use on
// the filesystem that holds path.
func AvailableBytes(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	blockSize := uint64(stat.Bsize)
	availableBlocks := uint64(stat.Bavail)
	if blockSize == 0 || availableBlocks > math.MaxUint64/blockSize {
		return 0, fmt.Errorf("filesystem free-space value overflows uint64")
	}
	return availableBlocks * blockSize, nil
}

// Rename is os.Rename. On Windows it retries while another process briefly
// holds either file open.
func Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// ReadFile is os.ReadFile. On Windows it retries while another process is
// briefly replacing the file.
func ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
