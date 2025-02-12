/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package malib

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// CheckForPanic returns a non-empty string when err contains a panic
//   - the string contains code location for the panic and the recovery
func CheckForPanic(err error) (panicString string) {
	if isPanic, stack, recoveryIndex, panicIndex := perrors.IsPanic(err); isPanic {
		var frames = stack.Frames()
		panicString = parl.Sprintf(
			"\nPANIC detected at %s\n"+
				"A Go panic may indicate a software problem\n"+
				"recovery was made at %s\n",
			frames[panicIndex].String(),
			frames[recoveryIndex].String(),
		)
	}
	return
}
