/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlfs

import (
	"os"
	"path/filepath"
)

// Entry is a file system entry that can be part of directory, symlink or other
type Entry struct {
	Base        string // we can have entries with FileInfo nil, so base path must be stored
	os.FileInfo        // info.Name() has basename, os.FileInfo: interface
	Err         error  // a symlink may fail to resolve
}

// NewEntry returns Entry or Directory
func NewEntry(path string, base string, sym func(string)) FSEntry {
	en := Entry{Base: base}
	en.FileInfo, en.Err = os.Lstat(path)
	if en.Err != nil {
		return &en
	}
	if en.Mode()&os.ModeSymlink != 0 {
		var symlinkTarget string
		if symlinkTarget, en.Err = filepath.EvalSymlinks(path); en.Err != nil {
			return &en
		}
		sym(symlinkTarget) // possibly add new root or restructure due to link to parent
	}
	if en.IsDir() {
		return NewDirectory(path, &en, sym)
	}
	return &en
}

// SafeName gets filepath.Base() in a safe way
func (en *Entry) SafeName() string {
	if en.FileInfo != nil {
		return en.FileInfo.Name()
	}
	return en.Base
}

// Walk traverses a built root
func (en *Entry) Walk(path string, walkFunc filepath.WalkFunc) error {
	return walkFunc(path, en.FileInfo, en.Err)
}
