/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"
)

/*
Walk pfs.Walk traverses a directory hierarchy following symlinks.
Every entry in the hierarchy is provided exactly once to walkFn.
To identify circular symlinks, the hierarchy is first completely scanned.
golang has a similar built-in filepath.Walk that does not follow symlinks
*/
func Walk(root string, walkFn filepath.WalkFunc) (err error) {
	tree := NewTree()
	var symlinkTargets []string
	if symlinkTargets, err = tree.AddRoot(root); err != nil {
		return
	}
	for _, symlinkTarget := range symlinkTargets {
		tree.ResolveSymlink(symlinkTarget)
	}

	if err = tree.Walk(walkFn); err != nil && err != filepath.SkipDir {
		return err
	}
	return nil
}
