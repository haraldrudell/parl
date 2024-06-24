/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

import (
	"os"

	"github.com/haraldrudell/parl/perrors"
)

var _ = os.WriteFile

// WriteNewFile writes data to filename without overwriting anything
//   - similar to [os.WriteFile] but does not overwrite and removes on error
//   - filename: filename, default current directory
//   - if err is non-nil, the file was not created or was removed due to error
//   - file must not exist
//   - file is created with ur permissions
func WriteNewFile(filename string, data []byte) (err error) {
	var osFile *os.File
	if osFile, err = NewFile(filename); err != nil {
		return
	}
	// close, delete if error
	defer CloseRm(filename, osFile, &err)

	if _, err = osFile.Write(data); err != nil {
		err = perrors.ErrorfPF("File.Write %w", err)
	}

	return
}
