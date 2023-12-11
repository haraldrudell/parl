/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/haraldrudell/parl/perrors"
)

type DeferringDirEntry struct {
	abs string
}

var _ fs.DirEntry = &DeferringDirEntry{}

// NewDeferringDirEntry returns [fs.DirEntry] with [fs.FileInfo] deferred
//   - abs is absolute path that maybe does not exist
func NewDeferringDirEntry(abs string) (entry *DeferringDirEntry) { return &DeferringDirEntry{abs: abs} }

// Info returns the FileInfo for the file or subdirectory described by the entry
//   - thread-safe, does [os.Lstat] every time
func (e *DeferringDirEntry) Info() (fileInfo fs.FileInfo, err error) {
	if fileInfo, err = os.Lstat(e.abs); err != nil {
		err = perrors.ErrorfPF("os.Lstat %w", err)
	}

	return
}

// IsDir reports whether the entry is a directory, ie. ModeDir bit being set
//   - may panic unless Info has successfully completed
func (e *DeferringDirEntry) IsDir() (isDir bool) {
	var fileInfo, err = e.Info()
	if err != nil {
		panic(err)
	}
	isDir = fileInfo.IsDir()

	return
}

// base name of the file
func (e *DeferringDirEntry) Name() (base string) { return filepath.Base(e.abs) }

// Type returns ModeType bits, 0 for regular file
//   - ModeDir ModeSymlink ModeNamedPipe ModeSocket ModeDevice ModeCharDevice ModeIrregular
func (e *DeferringDirEntry) Type() (modeType fs.FileMode) {
	var fileInfo, err = e.Info()
	if err != nil {
		panic(err)
	}
	modeType = fileInfo.Mode().Type()

	return
}
