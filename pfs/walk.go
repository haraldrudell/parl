/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/haraldrudell/parl"
)

var St = parl.Debug

// Walk pfs.Walk traverses a directory hierarchy following any symlinks
//   - every entry in the hierarchy is provided exactly once to walkFn
//   - to identify circular symlinks, the hierarchy is first completely scanned
//   - — this builds an in-memory representation of all files and directories
//   - — a complete scan is required toresolve nesting among symlinks
//   - the Go standard library [filepath.Walk] does not follow symlinks
//   - walkFn receives each entry with paths beginning similar to the path
//     provided to Walk
//   - — WalkFn is: func(path string, info fs.FileInfo, err error) error
//   - — path may be implicitly relative to current directory: “subdir/file.txt”
//   - — may have no directory part: “README.css”
//   - — may be relative: “../file.txt”
//   - — may contain symlinks, unnecessary “.” and “..”, and other problems
//   - — pfs.AbsEval returns an absolute, clean, symlink or non-symlink path
//   - — walkFn may receive an error if encountered while scanning a path
//   - — if walkFn receives an error, info may be nil
//     walkFn can choose to ignore, skip or return the error
//   - walkFn may return filepath.SkipDir to skip entering a certain directory
//   - walkFn may return filepath.SkipAll to end file system scanning
//   - Walk does not return filepath.SkipDir or filepath.SkipAll errors
func Walk(rootPath string, walkFn filepath.WalkFunc) (err error) {
	if parl.IsThisDebug() {
		dir, _ := os.Getwd()
		St("Walk: %q cd: %q", rootPath, dir)
	}

	// tree is a file system scan consisting of one or more roots
	var tree = NewTree(walkFn)

	// scan files and directories in the first root and any additional
	// encountered roots
	//	- if a symlink points outside the original root directory, this creates
	//		an additional root
	//	- if a symlink points above an existing root, this causes a rescan
	//		from the new superior root
	if err = tree.ScanRoots(rootPath); err != nil {
		return
	}

	// walk the tree
	if err = tree.Walk(); err != nil {
		if errors.Is(err, filepath.SkipDir) || errors.Is(err, filepath.SkipAll) {
			err = nil
		}
	}

	return
}
