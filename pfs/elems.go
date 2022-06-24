/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"
	"strings"

	"github.com/haraldrudell/parl/perrors"
)

var separator = string(filepath.Separator)

// Elems splits a path into a list of directory names and base filename part.
// if path is absolute, dirs[0] is "/".
// if there is no separator in path, dirs is empty.
// if path is empty string, dirs is empty and file is empty string.
func Elems(path string) (dirs []string, file string) {

	// first extract filename
	var path0 string
	if path0, file = filepath.Split(path); path0 == "" {
		return // no directory part exit
	}

	var path1 string
	var dir string
	for {

		// strip trailing separator
		path1 = strings.TrimSuffix(path0, separator)
		if path1 == path0 {
			panic(perrors.Errorf("pfs.Elems: missing trailing separator: %q separator: %q from: %q", path0, separator, path))
		}
		if path1 == "" {
			dirs = append([]string{path0}, dirs...)
			return // absolute path end exit
		}

		// extract last directory
		path1, dir = filepath.Split(path1)
		/*
			if path1 == path0 {
				dirs = append([]string{path0}, dirs...)
				return // no change in directory: at top return
			}
		*/
		dirs = append([]string{dir}, dirs...)
		if path1 == "" {
			return // end of a relative path exit
		}
		path0 = path1
	}
}
