/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"github.com/haraldrudell/parl/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

// IsPanic determines if err is the result of a panic. Thread-safe
// The err is an error value that may have an error chain and needs to have a stack trace
// captured in deferred recovery code arresting the panic.
//   - isPanic indicates if the error stack was captured during a panic
//   - stack[recoveryIndex] is the code line of the deferred function containing recovery invocation
//   - stack[panicIndex] is the code line causing the panic
//   - perrors.Short displays the error message along with the code location raising panic
//   - perrors.Long displays all available information about the error
func IsPanic(err error) (isPanic bool, stack pruntime.StackSlice, recoveryIndex, panicIndex int) {
	stack0 := errorglue.GetInnerMostStack(err)
	if len(stack0) == 0 {
		return // have no stack, assume not a panic
	}
	isPanic, recoveryIndex, panicIndex = errorglue.Indices(stack0)
	stack = stack0
	return
}
