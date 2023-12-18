/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"path/filepath"

	"github.com/haraldrudell/parl/perrors"
)

// Abs ensures a file system path is fully qualified.
// Abs is single-return-value and panics on troubles
func Abs(dir string) (out string) {
	var err error
	if out, err = filepath.Abs(dir); err != nil {
		panic(perrors.Errorf("filepath.Abs: '%w'", err))
	}
	return
}
