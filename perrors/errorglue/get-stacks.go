/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"slices"

	"github.com/haraldrudell/parl/pruntime"
)

// GetStacks gets a slice of all stack traces, oldest first
func GetStacks(err error) (stacks []pruntime.Stack) {
	for err != nil {
		if e, hasStack := err.(ErrorCallStacker); hasStack {
			var stack = e.StackTrace()
			// each stack encountered is older than the previous
			// store newest first
			stacks = append(stacks, stack)
		}
		err, _, _ = Unwrap(err)
	}
	// oldest first
	slices.Reverse(stacks)

	return
}
