/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import "io/fs"

// FSEntry represents a branch in a file system hierarchy. May be:
//   - an Entry, ie. a file and not a directory
//   - a Directory containing additional file-system entries
//   - a Root representing a file-system traversal entry point
type FSEntry interface {
	// Name is basename from fs.DirEntry “file.txt”
	Name() (base string)
	// IsDir returns whether the file-system entry is a directory containing files
	// as opposed to a file
	IsDir() (isDir bool)
	// IsSymlink returns whether the file-system entry is a symbolic link
	// that may create a new root
	//	- a symlink is not a directory, but may point to a directory that
	//		may become a new root
	IsSymlink() (isSymlink bool)
	// Walks returns data required for invoking filepath.WalkFunc
	Walks() (info fs.FileInfo, err error)
	// SetError stores an error encountered during scan for symlinks
	//	- it is provided to filepath.WalkFunc during the walk phase
	SetError(err error)
}

// GetFSEntry returns an FSentry corresponding to path
//   - the returned FSEntry may be a directory or a file-type entry
//   - path is provided path, possibly relative
//   - base is basename in case fileInfo cannot be retrieved
//   - sym receives a callback if the FSEntry is a symbolic link
func GetFSEntry(path string, base string) (entry FSEntry, err error) {

	// try to obtain FileInfo
	var entryImpl = NewEntry(base)
	entry = entryImpl
	if err = entryImpl.FetchFileInfo(path); err != nil {
		return // FileInfo failed
	}

	// directory case
	if entryImpl.IsDir() {
		entry = NewDirectory(entryImpl)
	}

	return
}
