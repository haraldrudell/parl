/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"

	"github.com/haraldrudell/parl/perrors"
)

// Root2 is a file system hierarchy
type Root2 struct {
	// path as provided that may be easier to read
	//   - may be implicitly relative to current directory: “subdir/file.txt”
	//   - may have no directory or extension part: “README.css” “z”
	//   - may be relative: “../file.txt”
	//   - may contain symlinks, unnecessary “.” and “..” or
	//		multiple separators in sequence
	//	- may be empty string for current working directory
	ProvidedPath string
	// Abs is required for relating roots
	//	- equivalent of Path: absolute, symlink-free, clean
	//	- may not exist
	Abs string
}

var _ = filepath.Clean

// NewRoot returns a file-system starting point for traversal
//   - path is as-provided path that may be:
//   - — absolute or relative
//   - — unclean
//   - — contain symlinks
func NewRoot2(path string) (root2 *Root2) { return &Root2{ProvidedPath: path} }

func NewAbsRoot2(abs string) (root2 *Root2) { return &Root2{ProvidedPath: abs, Abs: abs} }

// Load obtain the absolute, symlink-resolved path for the root
func (r *Root2) Load() (err error) {
	if r.Abs != "" {
		return // abs already present return
	}
	// absolute version of providedPath
	var abs string
	if abs, err = filepath.Abs(r.ProvidedPath); perrors.IsPF(&err, "filepath.Abs %w", err) {
		// Abs returns error if working directory cannot be determined
		return // error return
	} else if r.Abs, err = filepath.EvalSymlinks(abs); perrors.IsPF(&err, "filepath.EvalSymlinks %w", err) {
		// EvalSymlinks returns error if readlink fails
		return // error return
	}

	return // has r.Abs return
}
