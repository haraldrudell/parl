/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"errors"
	"io/fs"

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

// IsDirectory determines if path exists, is directory or another type of file-system entry
//   - path may be relative, contain symlinks or be unclean
//   - isDirectory is true if path exists and is directory.
//     Non-existing path is not an error.
//   - if IsDirectoryNonExistentIsError is present,
//     non-existing path returns error
//   - if IsDirectoryNotDirIsError is present,
//     a file-system entry that is not directory returns error
//   - flags is bitfield: IsDirectoryNonExistentIsError | IsDirectoryNotDirIsError
//   - symlinks are followed [os.Stat]
func IsDirectory(path string, flags IsDirectoryArg) (isDirectory bool, err error) {
	var fileInfo fs.FileInfo
	if fileInfo, err = Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
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
