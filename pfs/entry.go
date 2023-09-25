/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"os"

	"github.com/haraldrudell/parl/perrors"
)

// Entry is a file system entry that is not a directory
//   - can be part of directory, symlink or other
type Entry struct {
	// we can have entries with FileInfo nil, so base path must be stored
	base string
	// info.Name() has basename, os.FileInfo: interface
	//	- may be nil
	//	- Name() Size() Mode() ModTime() IsDir() Sys()
	os.FileInfo
	// if an error is encountered during scanning for symlinks,
	// that error is stored here, then presented to WalkFn
	// during the later walk phase
	err error
}

// NewEntry returns Entry or Directory
//   - path is absolute path to this file
func NewEntry(base string) (entry *Entry) {
	return &Entry{base: base}
}

func (e *Entry) FetchFileInfo(path string) (err error) {
	// Lstat describes a symlink, not what a symlink points to
	e.FileInfo, err = os.Lstat(path)
	perrors.Is(&err, "os.Lstat %w", err)
	return
}

func (e *Entry) IsSymlink() (isSymlink bool) {
	if f := e.FileInfo; f != nil {
		isSymlink = f.Mode()&os.ModeSymlink != 0
	}
	return
}

// Name gets filepath.Base() in a safe way
//   - Name() fails if FileInfo si not available
func (e *Entry) Name() (base string) {
	if e.FileInfo != nil {
		base = e.FileInfo.Name()
		return
	}
	base = e.base
	return
}

// Walk traverses which for a file is only the file itself
func (e *Entry) Walks() (info fs.FileInfo, err error) {
	info = e.FileInfo
	err = e.err
	return
}

// SetError stores an error encountered during scan for symlinks
//   - it is provided to filepath.WalkFunc during the walk phase
func (e *Entry) SetError(err error) {
	if e.err == nil {
		e.err = err
	} else {
		e.err = perrors.AppendError(e.err, err)
	}
}
