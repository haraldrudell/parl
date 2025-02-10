/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"errors"
	"io/fs"
	"os"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// creates a non-existing file, otherwise errors
	OpenFlagsMustNotExist = os.O_RDWR | // open the file read-write.
		os.O_CREATE | // create a new file if none exists.
		os.O_EXCL // used with O_CREATE, file must not exist.
		// read-and-writable by user, group-other no access
	PermUserReadWrite fs.FileMode = 0600
)

// OpenNew creates a file for writing, thread-safe across processes
//   - outfile: path “/file.txt”
//   - perms: default permissions: urw-------
//   - osFile: set if file was created, must then be closed
//   - isExist: it failed because the file already existed
func CreateFile(outfile string, perms ...fs.FileMode) (osFile *os.File, isExist bool, err error) {
	var perm fs.FileMode
	if len(perms) > 0 {
		perm = perms[0]
	} else {
		perm = PermUserReadWrite
	}
	if osFile, err = os.OpenFile(outfile, OpenFlagsMustNotExist, perm); perrors.IsPF(&err, "OpenFile %w", err) {
		isExist = errors.Is(err, os.ErrExist)
	}
	return
}
