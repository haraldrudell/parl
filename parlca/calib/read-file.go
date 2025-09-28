/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package calib

import (
	"os"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/punix"
)

// ReadFile is [os.ReadFile] allowing for file to be missing
func ReadFile(filename string, allowNotFound bool) (byts []byte, err error) {
	if byts, err = os.ReadFile(filename); err != nil {
		if allowNotFound && punix.IsENOENT(err) {
			err = nil
			// cert file does not exist: byts == nil, err == nil
			return
		}
		perrors.IsPF(&err, "os.ReadFile %w", err)
	}
	return
}
