/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package parlos provides simplified functions related to the os package
package pos

import (
	"os"
	"path/filepath"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfs"
)

// ParentDir gets absolute path of executable parent directory
func ParentDir() (dir string) {
	return filepath.Dir(ExecDir())
}

// ExecDir gets abolute path to directory where executable is located
func ExecDir() (dir string) {
	var executable string
	var err error
	if executable, err = os.Executable(); err != nil {
		panic(perrors.Errorf("os.Executable: %w", err))
	}
	dir = pfs.Abs(filepath.Dir(executable))
	return
}
