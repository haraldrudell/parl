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

// Exists determines if a path exists
// If the path exists, fileInfo is non-nil
// if the path does not exist, fileInfo is nil
// panic on troubles
func Exists(path string) (fileInfo fs.FileInfo /* interface */) {
	var err error
	fileInfo, err = os.Stat(path)
	if err == nil {
		return // does exist: fileInfo
	}
	if os.IsNotExist(err) {
		return // does not exist: nil
	}
	panic(perrors.Errorf("os.Stat: '%w'", err))
}
