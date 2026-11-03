// Package osutil holds the few operating-system calls whose behavior differs
// between Unix and Windows, so the rest of the code stays portable.
//
// Each function has a Unix implementation (osutil_unix.go, the behavior the
// code had before this package existed) and a Windows one (osutil_windows.go).
package osutil

import "errors"

// ErrLocked reports that another process already holds an exclusive lock.
var ErrLocked = errors.New("file is locked by another process")
