/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"slices"

	"github.com/haraldrudell/parl/perrors/panicdetector"
	"github.com/haraldrudell/parl/pruntime"
)

// FirstPanicStack checks all stack traces, oldest first, for
// a panic and returns:
//   - stack:
//   - — the panic stack if any
//   - — otherwise, the oldest stack if any
//   - — otherwise nil
//   - isPanic true: a panic stack was found
//   - — recoveryIndex panicIndex are valid
//   - numberOfStacks: the number of stacks until the panic stack including it,
//     or the total number of stacks
//   - errorWithStack:
//   - — if panic, the error providing the oldest panic stack
//   - — otherwise, the oldest error with a stack
//   - — otherwise, the oldest error
//   - — otherwise nil
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
	isPanic bool, stack pruntime.Stack, recoveryIndex, panicIndex int,
	numberOfStacks int,
	errorWithStack error,
) {

	// any panic stack need to be associated with its exact error
	//	- normally, the returned stack is from any older error in the chain
	//	- therefore, the err’s error chain must be traversed
	//	- the error chain is a directed single-linked list newest error first
	//	- —
	//	- if there is no panic, the oldest stack trace is what will be displayed, ie. the source of the error
	//	- a panic can be in a newer stack trace if an error with a stack trace is provided to panic
	//	- therefore, the oldest panic stack trace should be provided

	// errs is the expanded error chain
	//	- err first that is newest, then older errrors
	var errs []error
	for e := err; e != nil; e, _, _ = Unwrap(e) {
		errs = append(errs, e)
	}
	// errs now begin with oldest error, ending with err
	slices.Reverse(errs)

	// oldestStack contains the oldest stack trace
	//	- if there is no panic, this should be returned
	var oldestStack pruntime.Stack
	// the oldest error, or
	// if a stack was found, the oldest error with a stack
	var oldestErr error

	// scan for panic-stack, oldest error first
	for _, e := range errs {

		// check if e contains a stack trace
		if errWithStack, ok := e.(ErrorCallStacker); ok {
			// save stack
			stack = errWithStack.StackTrace()
			if oldestStack == nil {
				// store oldest stack
				oldestStack = stack
				// store oldest error or oldest error with stack
				oldestErr = e
			}
			// store the error providing stack
			errorWithStack = e
		} else {
			continue // an error without stack
		}
		// increase number of stack until panic,
		// or total number of stacks
		numberOfStacks++

		// check if the stack trace has a panic
		isPanic, recoveryIndex, panicIndex = panicdetector.Indices(stack)
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
