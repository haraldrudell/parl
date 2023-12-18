/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"os"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// if path does not exist, return the error [IsDirectory]
	IsDirectoryNonExistentIsError IsDirectoryArg = 1 << iota
	// if existing path is not directory, return an error [IsDirectory]
	//	- symlinks are followed, so a symlink pointing ot directory is ok
	IsDirectoryNotDirIsError
)

// bitfield argument to [IsDirectory]
//   - IsDirectoryNonExistentIsError IsDirectoryNotDirIsError
type IsDirectoryArg uint8

// IsDirectory determines if path exists, is a directory or other entry
//   - flags is bitfield
//   - flags: IsDirectoryNonExistentIsError
//   - flags: IsDirectoryNotDirIsError
func IsDirectory(path string, flags IsDirectoryArg) (isDirectory bool, err error) {
	var fileInfo fs.FileInfo
	if fileInfo, err = Stat(path); err != nil {
		if os.IsNotExist(err) {
			if flags&IsDirectoryNonExistentIsError == 0 {
				err = nil
			}
		}
		return // stat error return, possibly ignored
	}
	if isDirectory = fileInfo.IsDir(); !isDirectory && IsDirectoryNotDirIsError != 0 {
		err = perrors.Errorf("Not directory: %s", path)
	}

	return
}
