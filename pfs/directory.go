/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"
)

// Directory is a file system entry with children
type Directory struct {
	Entry
	Children []FSEntry
}

// NewDirectory instantiates FSEntry that can have children
func NewDirectory(entryImpl *Entry) (d *Directory) {
	return &Directory{Entry: *entryImpl}
}

// FetchChildren reads the directory and populates the Children field
//   - errors are stored using SetError
func (d *Directory) FetchChildren(path string) (children []FSEntry, paths []string, err error) {

	// read the directory
	var entryNames []string
	if entryNames, err = readDir(path); err != nil {
		d.SetError(err)
		return
	}

	// create entry instances for children
	children = make([]FSEntry, len(entryNames))
	paths = make([]string, len(entryNames))
	for i, entryName := range entryNames {
		var path = filepath.Join(path, entryName)
		paths[i] = path
		// error is stored in child
		var entry FSEntry
		entry, err = GetFSEntry(path, entryName)
		children[i] = entry
		if err != nil {
			entry.SetError(err)
		}
	}

	d.Children = children

	return
}
