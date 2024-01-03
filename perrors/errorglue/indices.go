/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"strings"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	runtimePrefix = "runtime."
)

// Indices examines a stack to see if it includes a panic. Thread-safe
//   - isPanic is true if the stack includes a panic
//   - stack[recoveryIndex] is the code line of the deferred function containing recovery invocation
//   - stack[panicIndex] is the code line causing the panic
func Indices(stack pruntime.StackSlice) (isPanic bool, recoveryIndex, panicIndex int) {
	found := 0
	stackLength := len(stack)
	pd := panicDetectorOne
	for i := 0; i < stackLength; i++ {
		funcName := stack[i].FuncName
		if i > 0 && funcName == pd.runtimeDeferInvokerLocation {
			recoveryIndex = i - 1
			found++
			if found == 2 {
				break
			}
		}
		if i+1 < stackLength && funcName == pd.runtimePanicFunctionLocation {

			// scan for end of runtime functions
			for panicIndex = i + 1; panicIndex+1 < stackLength; panicIndex++ {
				if !strings.HasPrefix(stack[panicIndex].FuncLine(), runtimePrefix) {
					break // this frame not part of the runtime
				}
			}

			found++
			if found == 2 {
				break
			}
		}
	}
	isPanic = found == 2
	return
}

// FirstPanicStack checks all stack traces, oldest first, for
// a panic and returns:
//   - — the panic stack if any
//   - — otherwise, the oldest stack if any
//   - this enables consumer to print:
//   - — panic location if any
//   - — otherwise, oldest error location if any
//   - — otherwise no location
//   - err is an error chain that may contain stack traces and be nil
//   - if any stack trace has a panic:
//   - — isPanic is true
//   - — stack is the oldest stack with a panic
//   - — stack[panicIndex] is the code line causing the oldest panic
//   - — stack[recoveryIndex] is the first code line in the deferred sequence.
//     Typically the last line of the executing function
//   - — numberOfStacks is 1 for the panic stack and added the number of stacks that are older
//   - — errorWithStack is the error with panic, typically from [parl.Recover]
//   - if isPanic is false:
//   - — panicIndex recoveryIndex are zero
//   - — stack and errorWithStack is the oldest stack if any
//     -— numberOfStacks is total number of stacks in the error chain
//   - this must be done beacuse maybe panic() was
//     provided an error with stack trace
//   - on err nil: all values false, nil or zero
func FirstPanicStack(err error) (
	isPanic bool, stack pruntime.StackSlice, recoveryIndex, panicIndex int,
	numberOfStacks int,
	errorWithStack error,
) {

	// need to associate any panic stack with its exact error
	// must traverse errors from oldest to newest
	//	- an error chain is a single-linked list newest first
	//	- this list must be reversed
	var errs []error
	for e := err; e != nil; e = errors.Unwrap(e) {
		errs = append(errs, e)
	}

	// if there is no panic,
	// the oldest stack should be returned
	//	- keep track of it here
	var oldestStack pruntime.StackSlice
	var oldestErr error

	// scan for panic-stack, oldest error first
	for _, e := range errs {
		if errWithStack, ok := e.(ErrorCallStacker); ok {
			stack = errWithStack.StackTrace()
			if oldestStack == nil {
				oldestStack = stack
				oldestErr = e
			}
			errorWithStack = e
		} else {
			continue // an error without stack
		}
		numberOfStacks++

		// check if the stack trace has a panic
		isPanic, recoveryIndex, panicIndex = Indices(stack)
		if !isPanic {
			continue // this error did not have a panic stack-trace
		}

		return // found panic return
	}
	// no panic stack was found
	stack = oldestStack
	errorWithStack = oldestErr

	return // no panic-stack return

}
