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

func Stat(path string) (fileInfo fs.FileInfo, err error) {
	if fileInfo, err = os.Stat(path); err != nil {
		err = perrors.Errorf("os.Stat: ‘%w’", err)
	}
	return
}
