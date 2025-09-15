/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"github.com/haraldrudell/parl/pruntime"
)

// GetInnermostStack gets the oldest stack trace in the error chain
// or nil if no stack trace is present
func GetInnermostStack(err error) (stack pruntime.Stack) {

	// find the innermost implementation of ErrorCallStacker interface
	var e ErrorCallStacker
	for ; err != nil; err, _, _ = Unwrap(err) {
		if ecs, ok := err.(ErrorCallStacker); ok {
			e = ecs
		}
	}
	if e == nil {
		return // no implementation found
	}
	stack = e.StackTrace()

	return
}
