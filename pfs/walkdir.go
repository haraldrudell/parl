/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"os"
	"path/filepath"
	"sort"
)

// Directory is a file system entry with children
type Directory struct {
	Entry
	Children []FSEntry
}

// NewDirectory instantiates FSEntry that can have children
func NewDirectory(path string, entry *Entry, sym func(string)) FSEntry {
	dir := Directory{Entry: *entry}
	var entryNames []string
	if entryNames, dir.Err = readDir(path); dir.Err != nil {
		return &dir
	}
	dir.Children = make([]FSEntry, len(entryNames))
	for i, entryName := range entryNames {
		dir.Children[i] = NewEntry(filepath.Join(path, entryName), entryName, sym)
	}
	return &dir
}

// Walk traverses a built root
func (dir *Directory) Walk(path string, walkFunc filepath.WalkFunc) error {
	if err := walkFunc(path, dir.FileInfo, dir.Err); err != nil {
		return err
	}
	for _, entry := range dir.Children {
		if err := entry.Walk(filepath.Join(path, entry.SafeName()), walkFunc); err != nil && err != filepath.SkipDir {
			return err
		}
	}
	return nil
}

// readDir reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDir(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}
