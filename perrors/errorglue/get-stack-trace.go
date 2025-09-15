/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"github.com/haraldrudell/parl/pruntime"
)

// GetStackTrace gets the last stack trace
func GetStackTrace(err error) (stack pruntime.Stack) {
	for ; err != nil; err, _, _ = Unwrap(err) {
		if ecs, ok := err.(ErrorCallStacker); ok {
			stack = ecs.StackTrace()
			return
		}
	}
	return
}
