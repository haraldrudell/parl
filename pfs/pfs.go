/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrEndListing = errors.New("endListing")

// Dirs retrieves absolute paths to all directories, while following symlinks, from initial dir argument.
// callback: cb is 6–58% faster than slice, results are found faster, and it can be canceled midway.
// if callback blocks, not good…
func Dirs(dir string, callback ...func(dir string) (err error)) (dirs []string, err error) {
	var callback0 func(dir string) (err error)
	if len(callback) > 0 {
		callback0 = callback[0]
	}

	// make dir path absolute
	var dir0 string
	if dir0, err = filepath.Abs(dir); err != nil {
		return
	}

	// find directories to watch
	if err = Walk(dir0, func(path string, info os.FileInfo, err0 error) (err error) {
		if err0 != nil {
			return err0 // some error occured during Walk exit
		}
		if info.IsDir() {
			if callback0 != nil {
				if err = callback0(path); err != nil {
					return // callback error or abort exit
				}
			} else {
				dirs = append(dirs, path)
			}
		}
		return // good exit
	}); err == ErrEndListing {
		err = nil
	}
	return
}
