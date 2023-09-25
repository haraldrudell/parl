/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"github.com/haraldrudell/parl/perrors"
)

// Root is a file system hierarchy
type Root struct {
	// path as provided that may be easier to read
	//   - — may be implicitly relative to current directory: “subdir/file.txt”
	//   - — may have no directory part: “README.css”
	//   - — may be relative: “../file.txt”
	//   - — may contain symlinks and unnecessary “.” and “..”
	ProvidedPath string
	// cleaned equivalent of Path
	//	- absolute, no symlinks, no unnecessary “.” and “..” or other problems
	absPath string
	// FSEntry represents the file-system location of this root
	//	- Directory or Entry
	//	- SafeName() IsDir()
	FSEntry
}

// NewRoot returns a file-system starting point for traversal
//   - path may be absolute or relative path
func NewRoot(path string) (root *Root) {
	return &Root{ProvidedPath: path}
}

func (r *Root) Init() (absPath string, rootEntry FSEntry, err error) {
	if r.FSEntry != nil {
		panic(perrors.NewPF("invoked more than once"))
	}

	// retrieve absolute and clean path
	if absPath, err = AbsEval(r.ProvidedPath); err != nil {
		return
	}
	r.absPath = absPath

	// create root entry
	// error is stored in r.FSEntry
	if rootEntry, err = GetFSEntry(r.ProvidedPath, ""); err != nil {
		rootEntry.SetError(err)
		err = nil
	}
	r.FSEntry = rootEntry

	return
}

func (r *Root) Abs() (abs string) {
	return r.absPath
}
