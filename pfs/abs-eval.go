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
	RetainSymlinks = false
)

// AbsEval returns an absolute path with resolved symlinks
//   - use of current directory to make path absolute
//   - “.” is returned for empty string
//   - if retainSymlinks is RetainSymlinks, symlinks are returned
//   - correct multiple Separators, eliminate “.”, eliminate inner and
//     inappropriate “..”, ensure platform Separator
func AbsEval(path string, retainSymlinks ...bool) (absPath string, err error) {
	if absPath, err = filepath.Abs(path); perrors.Is(&err, "filepath.Abs %w", err) {
		return
	}
	if len(retainSymlinks) == 0 || !retainSymlinks[0] {
		absPath, err = filepath.EvalSymlinks(absPath)
		perrors.Is(&err, "EvalSymlinks %w", err)
	} else {
		absPath = filepath.Clean(absPath)
	}

	return
}
