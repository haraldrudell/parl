/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
)

// GoResult makes a small fixed number of goroutines awaitable at
// minimum panic, error, bugs, wait and dead-lock troubles
//   - requires new-function invocation:
//   - — [NewGoResult] is the simplest, goroutines are awaited by [GoResult.ReceiveError]
//   - — [NewGoResult2] has error flag and goroutine counter
//   - — number of goroutines must be known at time of new, it dimensions the channel
//   - [GoResult.SendError](errp *error) deferrable, how goroutine sends results
//   - [GoResult.ReceiveError](errp *error, n ...int) (err error) how managing thread
//     receives goroutine exits
//   - [NewGoResult2] also has:
//   - — [GoResult.IsError]() (isError bool) true if any goroutine returned error
//   - — [GoResult.SetIsError]() sets the error flag manually
//   - — [GoResult.Remaining](add ...int) goroutine counter
//   - —
//   - Requirement: GoResult should be pass-by-value similar to channel, not as pointer
//   - Requirement: GoResult should be usable in variable declaration
//   - Requirement: two innermost types, simple and feature-rich implementations
//   - Requirement: the innermost concrete type could be chan or struct with multiple fields
//   - to avoid duplication when passed as function parameter, innermost type must be pointed to
//   - to support multiple innermost types, the type chain must include interface
//     which implictly is pointer
//   - to support interface, the GoResult type cannot be chan
//   - to be in a variable declaration, GoResult cannot be interface
//   - to minimize levels of indirection, GoResult is struct value with single interface field
//   - note: a type with methods cannot be pointer or interface-value
//   - because GoResult is used multiple times in go and defer statements, it must have identifier
//   - new-functions returns GoResult value with internal pointer
type GoResult struct{ goResult }

// GoResult implements Done(errp *error)
var _ Doner = &GoResult{}

// NewGoResult returns the minimum mechanic to make goroutines awaitable
//   - n: optional goroutine capacity, default 1
//   - — capacity ensures goroutines do not block on exit
//   - —
//   - a thread-creator provides GoResult to its goroutines in go statements making them awaitable
//   - each exiting thread sends an error value that may be nil
//   - mechanic is buffered channel
//
// Usage:
//
//	func someFunc(text string) (err error) {
//	  var g = parl.NewGoResult()
//	  go goroutine(text, g)
//	  defer g.ReceiveError(&err)
//	  …
//	func goroutine(text string, g parl.GoResult) {
//	  var err error
//	  defer g.Done(&err)
//	  defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)
//
//	  err = …
func NewGoResult(n ...int) (goResult GoResult) {
	return GoResult{goResult: newGoResultChan(n...)}
}

// NewGoResult2 is [NewGoResult] with additional error flag and goroutine counter
//   - n: optional goroutine capacity, default 1
//   - — capacity ensures goroutines do not block on exit
//   - —
//   - [GoResult.Remaining] a counter enabling waiting for fewer goroutines than n
//   - [GoResult.IsError] true if any goroutine exited with error or SetIsError was invoked
//   - [GoResult.SetIsError] sets error state regardless of goroutine exits
//   - a thread-creator provides GoResult to its goroutines in go statements making them awaitable
//   - each exiting thread sends an error value that may be nil
//   - mechanic is buffered channel
//
// Usage:
//
//	func someFunc(text string) (err error) {
//	  var g = parl.NewGoResult2()
//	  go goroutine(text, g)
//	  defer g.ReceiveError(&err)
//	  …
//	func goroutine(text string, g parl.GoResult) {
//	  var err error
//	  defer g.Done(&err)
//	  defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)
//
//	  err = …
func NewGoResult2(n ...int) (goResult GoResult) {
	return GoResult{goResult: newGoResultStruct(newGoResultChan(n...))}
}

// true if the GoResult is initialized by new-function
func (g GoResult) IsValid() (isValid bool) { return g.goResult != nil }

// Done indicates that this goroutine is exiting
//   - *errp nil: successful exit
//   - *errp non-nil: fatal error exit
//   - —
//   - deferrable
//   - Done makes a goroutine:
//   - — awaitable and
//   - — able to return error
//   - — other needs of a goroutine is to initiate and detect cancel and
//     submit non-fatal errors
func (g GoResult) Done(errp *error) {

	// get implementation
	var gp = g.goResult
	if gp == nil {
		panic(perrors.NewPF("uninitialized GoResult"))
	}
	NilPanic("errp", errp)

	g.done(*errp)
}

// ReceiveError is a deferrable function receiving error values from goroutines
//   - n: number of goroutines to wait for
//   - n missing: wait for all goroutines.
//     [NewGoResult]: the number provided to new-function
//   - [NewGoResult2]: if adds non-zero, wait for adds goroutines.
//     Otherwise, wait for number provided to new-function
//   - — if consumer uses Remaining and adds may be zero,
//     the output from Remaining must be provided to ReceiveError.
//   - errp: may be nil
//   - ReceiveError makes a goroutine:
//   - — awaitable and
//   - — able to return a fatal error
//   - — other needs of a goroutine is to initiate and detect cancel and
//     submit non-fatal errors
//   - GoRoutine should have enough capacity for all its goroutines
//   - — this prevents goroutines from blocking in channel send
//   - ReceiveError only panics from structural coding problems
//   - deferrable thread-safe
func (g GoResult) ReceiveError(errp *error, n ...int) (err error) {

	// get implementation
	var gp = g.goResult
	if gp == nil {
		panic(perrors.NewPF("uninitialized GoResult"))
	}

	// the error-receiving channel
	var ch = gp.ch()

	// get error count
	var remainingErrors int
	if len(n) > 0 {
		remainingErrors = n[0]
	} else if goResultStruct, _ := g.goResult.(*goResultStruct); goResultStruct != nil {
		remainingErrors = goResultStruct.remaining(0)
	} else {
		// default for simple GoResult: dimension of channel
		remainingErrors = cap(ch)
	}

	// await goroutine results
	for i := range remainingErrors {
		_ = i

		// blocks here
		//	- wait for a result from a goroutine
		var e = <-ch
		if e == nil {
			continue // good return: ignore
		}

		// goroutine exited with error
		// ensure e has stack
		e = perrors.Stack(e)
		// build error list
		err = perrors.AppendError(err, e)
	}

	// final action: update errp if present
	if err != nil && errp != nil {
		*errp = perrors.AppendError(*errp, err)
	}

	return
}

//   - available: the number of results that can be currently collected.
//     ie. len of the results channel which is
//     Done invocations yet to be collected by ReceiveError
//   - stillRunning [NewGoResult2] only: the number of created goroutines yet to invoke SendError.
//     That is cumulative adds less SendError invocations.
//     If adds is zero, the dimensioned capacity provided to new-function
//     less SendError invocations
//   - Thread-safe
//   - —
//   - stillRunning lack integrity with Remaining and Done invocations
//     compared to available
//   - — a parallel Remaining may increase stillRunning
//   - — a parallel Done may decrease stillRunning
func (g GoResult) Count() (available, stillRunning int) {

	// get implementation
	var gp = g.goResult
	if gp == nil {
		panic(perrors.NewPF("uninitialized GoResult"))
	}

	return g.count()
}

// IsError returns if any goroutine has returned an error
//   - only for [NewGoResult2]
func (g GoResult) IsError() (isError bool) { return g.doError() }

// SetIsError sets error state regardless of whether any goroutine has returned an error
//   - consumer will exhibit error behavior without fatal thread-exit
//   - only for [NewGoResult2]
func (g GoResult) SetIsError() { g.doError(true) }

// doError handles set and detect of errors
func (g GoResult) doError(setIsError ...bool) (isError bool) {

	// get implementation
	var goResultStruct, _ = g.goResult.(*goResultStruct)
	if goResultStruct == nil {
		if g.goResult == nil {
			panic(perrors.NewPF("uninitialized GoResult"))
		}
		panic(perrors.NewPF("NewGoResult does not provide IsError SetIsError: use NewGoResult2"))
	}

	// get set error
	var setError = len(setIsError) > 0 && setIsError[0]

	return goResultStruct.doError(setError)
}

// Remaining returns the number of goroutines that should be awaited
//   - add: optional add for count-based number of created goroutines
//   - — add cannot be negative
//   - adds: the cumulative number of add values provided
//   - — adds allow for not waiting on goroutines that were never created
//   - if adds is zero, ie. no add was ever provided, adds is the dimensioned
//     capacity provided to the new-function
//   - only for [NewGoResult2]
func (g GoResult) Remaining(add ...int) (adds int) {

	// get implementation
	var goResultStruct, _ = g.goResult.(*goResultStruct)
	if goResultStruct == nil {
		if g.goResult == nil {
			panic(perrors.NewPF("uninitialized GoResult"))
		}
		panic(perrors.NewPF("NewGoResult does not provide IsError SetIsError: use NewGoResult2"))
	}

	// get add
	var addValue int
	if len(add) > 0 {
		addValue = add[0]
	}
	if addValue < 0 {
		panic(perrors.ErrorfPF("negative add: %d", addValue))
	}

	return goResultStruct.remaining(addValue)
}

// String method always works
//   - “GoResult_nil”:
//     uninitialized, invalid GoResult nil or no new-function
//   - “goResult_len:0”: from NewGoResult
//   - — has channel capacity 1
//   - — len is whether any result is pending in the channel
//   - “goResult_adds:2_sends:1_ch:0(2)_isError:false” from NewGoResult2
//   - — Has buffered channel of certain capacity
//   - — remain is how many results remain to be read from channel
//   - — ch is how many items are pending in channel
//   - — isError is true if an error was read from a goroutine
//     or SetIsError was invoked
func (g GoResult) String() (s string) {
	if g.goResult == nil {
		return "GoResult_nil"
	}
	s = g.goResult.String()

	return
}
