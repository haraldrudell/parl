/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"io"
	"os"

	"github.com/haraldrudell/parl/perrors"
)

// IsEmptyDirectory checks if a directory is empty.
// ignoreENOENT if true, a non-exiting directory is ignored.
// pamnic on troubles
func IsEmptyDirectory(path string, ignoreENOENT ...bool) (isEmpty bool) {
	var ign bool
	if len(ignoreENOENT) > 0 {
		ign = ignoreENOENT[0]
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && ign {
			return false
		}
		panic(perrors.Errorf("os.Open: '%w'", err))
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(perrors.Errorf("file.Close: '%w'", err))
		}
	}()
	_, err = file.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			return true // empty directory
		}
		panic(perrors.Errorf("Readdirnames: '%w'", err))
	}
	return false // directory has files
}
