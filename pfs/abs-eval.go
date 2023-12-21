/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// do not evaluate symlinks argument to [AbsEval]
	RetainSymlinks = true
)

// AbsEval returns an absolute path with resolved symlinks
//   - if path is relative, process’ current directory is used to make path absolute.
//     path empty returns the process’ current directory as absolute evaled path
//   - absPath is absolute and only empty on error.
//   - absPath is clean:
//   - — no multiple Separators in sequence
//   - — no “.” path name elements
//   - — no infix or postfix “..” path name element or to path above root
//   - when evaluating symlinks, the path is verified to exist
//   - if retainSymlinks is RetainSymlinks, symlinks are returned
func AbsEval(path string, retainSymlinks ...bool) (absPath string, err error) {

	// ensure path is absolute
	if !filepath.IsAbs(path) {
		if absPath, err = filepath.Abs(path); perrors.IsPF(&err, "filepath.Abs %w", err) {
			return
		}
	} else {
		absPath = path
	}

	// evaluate symlink
	if len(retainSymlinks) == 0 || !retainSymlinks[0] {
		if absPath, err = filepath.EvalSymlinks(absPath); perrors.IsPF(&err, "EvalSymlinks %w", err) {
			return
		}
	} else {
		absPath = filepath.Clean(absPath)
	}

	return
}
