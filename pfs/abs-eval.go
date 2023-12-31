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
//   - if symlinks are evaluated, the path must exist
//   - non-existing path is checked using punix.IsENOENT(err)
//   - if path is relative, the process’ current directory is used to make path absolute.
//     path empty returns the process’ current directory as absolute evaled path
//   - absPath is absolute and only empty on error
//   - absPath is clean:
//   - — no multiple Separators in sequence
//   - — no “.” path name elements
//   - — no infix or postfix “..” path name element or to path above root
//   - when evaluating symlinks, the path is verified to exist
//   - if retainSymlinks is RetainSymlinks, symlinks are returned
//   - errors originate from:
//   - — [os.Lstat] [os.Readlink] [os.Getwd] or
//   - — infix path-segment not directory or
//   - — encountering more than 255 symlinks
func AbsEval(path string, retainSymlinks ...bool) (absPath string, err error) {
	var didClean bool

	// ensure path is absolute
	if didClean = !filepath.IsAbs(path); didClean {
		// errors from [os.Getwd]
		if absPath, err = filepath.Abs(path); perrors.IsPF(&err, "filepath.Abs %w", err) {
			return
		}
	} else {
		absPath = path
	}

	// evaluate symlink
	if len(retainSymlinks) == 0 || !retainSymlinks[0] {
		// errors from [os.Lstat] [os.Readlink] or
		//	- infix path segment not directory
		//	- or encountering more than 255 symlinks
		if absPath, err = filepath.EvalSymlinks(absPath); perrors.IsPF(&err, "EvalSymlinks %w", err) {

			// // *errorglue.errorStack *fmt.wrapError *fs.PathError syscall.Errno
			// parl.Log("filepath.EvalSymlinks ENOT error chanin: %s", errorglue.DumpChain(err))
			// // os.IsNotExist fails: false
			// parl.Log("os.IsNotExist fails: %t", os.IsNotExist(err))
			// if errnoValue := punix.Errno(err); errnoValue != 0 {
			// 	// errno: ENOENT 2 0x2
			// 	parl.Log(punix.ErrnoString("errno", errnoValue))
			// }
			// // punix.IsENOENT: true
			// parl.Log("punix.IsENOENT: %t", punix.IsENOENT(err))

			return
		}
	} else if !didClean {
		absPath = filepath.Clean(absPath)
	}

	return
}
