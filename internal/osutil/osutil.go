// Package osutil holds the few operating-system calls whose behavior differs
// between Unix and Windows, so the rest of the code stays portable.
//
// Each OS call has a Unix implementation (osutil_unix.go, the behavior the
// code had before this package existed) and a Windows one (osutil_windows.go).
// SQLiteFileURI is shared: it differs only in how it treats drive letters.
package osutil

import (
	"errors"
	"net/url"
	"path/filepath"
)

// ErrLocked reports that another process already holds an exclusive lock.
var ErrLocked = errors.New("file is locked by another process")

// SQLiteFileURI returns a SQLite "file:" URI for path with rawQuery appended,
// such as "mode=ro". On Windows a drive path like C:\data\messages.db must
// become file:///C:/data/messages.db. Without the slash before the drive
// letter, SQLite reads "C:" as the URI's authority and refuses to open the
// file, which modernc.org/sqlite reports as "out of memory". On other
// systems the result is unchanged from building the URL directly.
func SQLiteFileURI(path, rawQuery string) string {
	uriPath := filepath.ToSlash(path)
	if volume := filepath.VolumeName(path); len(volume) == 2 && volume[1] == ':' {
		uriPath = "/" + uriPath
	}
	return (&url.URL{Scheme: "file", Path: uriPath, RawQuery: rawQuery}).String()
}
