/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io/fs"
	"os"

	"github.com/haraldrudell/parl/perrors"
)

var _ fs.FileInfo
var _ fs.FileMode
var _ = fs.FileMode.IsDir
var _ = fs.FileMode.Type

// AddDirEntry returns an [fs.DirEntry] with [fs.FileInfo] available
//   - error if [os.Lstat] fails
func AddDirEntry(abs string) (dirEntry fs.DirEntry, err error) {
	var fileInfo fs.FileInfo
	if fileInfo, err = os.Lstat(abs); perrors.IsPF(&err, "os.Lstat %w", err) {
		return
	}
	dirEntry = fs.FileInfoToDirEntry(fileInfo)

	return
}

// AddDirEntry returns an [fs.DirEntry] for the target of a symbolic link
func AddStatDirEntry(abs string) (dirEntry fs.DirEntry, err error) {
	var fileInfo fs.FileInfo
	if fileInfo, err = os.Stat(abs); perrors.IsPF(&err, "os.Stat %w", err) {
		return
	}
	dirEntry = fs.FileInfoToDirEntry(fileInfo)

	return
}
