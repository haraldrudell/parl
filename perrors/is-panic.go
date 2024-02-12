/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

// IsPanic determines if err is the result of a panic. Thread-safe
//   - isPanic is true if a panic was detected in the inner-most stack trace of err’s error chain
//   - err must have stack trace from [perrors.ErrorfPF] [perrors.Stackn] or similar function
//   - stack[recoveryIndex] is the code line of the deferred function containing recovery invocation
//   - stack[panicIndex] is the code line causing the panic
//   - [perrors.Short] displays the error message along with the code location raising panic
//   - [perrors.Long] displays all available information about the error
func IsPanic(err error) (isPanic bool, stack pruntime.Stack, recoveryIndex, panicIndex int) {
	isPanic, stack, recoveryIndex, panicIndex, _, _ = errorglue.FirstPanicStack(err)
	return
}
