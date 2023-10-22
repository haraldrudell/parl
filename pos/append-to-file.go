/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"os"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// when created, output file permissions is user-read/write
	FilePermUrw os.FileMode = 0o600 // rw- --- ---
	// flags for os.OpenFile: must be new, write-only
	openFlagsCreateOrAppend = os.O_CREATE | os.O_APPEND | os.O_WRONLY
)

func AppendToFile(filename string) (osFile *os.File, err error) {
	if osFile, err = os.OpenFile(filename, openFlagsCreateOrAppend, FilePermUrw); err != nil {
		err = perrors.ErrorfPF("OpenFile: %w", err)
	}

	return
}
