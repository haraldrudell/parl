/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

/*
Package pfs provides file-system related functions
*/
package pfs

import (
	"os"
	"path/filepath"
)

// AbsEval makes path absolute and resolves symlinks
func AbsEval(path string) (p string, err error) {
	if p, err = filepath.Abs(path); err == nil {
		p, err = filepath.EvalSymlinks(p)
	}
	return
}

// Dirs retrieves absolute paths to all directories, while following symlinks, from initial dir argument
func Dirs(dir string) (dirs []string, err error) {

	// make dir path absolute
	var dir0 string
	if dir0, err = filepath.Abs(dir); err != nil {
		return
	}

	// find directories to watch
	err = Walk(dir0, func(path string, info os.FileInfo, err0 error) (err error) {
		if err0 != nil {
			return err0
		}
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return
	})
	return
}
