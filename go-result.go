/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

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

// true if the GoResult is initialized by new-function
func (g GoResult) IsValid() (isValid bool) { return g.goResult != nil }

// String method always works
//   - “GoResult_nil”:
//     uninitialized, invalid GoResult nil or no new-function
//   - “goResult_len:0”: from NewGoResult
//   - — has channel capacity 1
//   - — len is whether any result is pending in the channel
//   - “goResult_remain:1_ch:0(1)_isError:false” from NewGoResult2
//   - — Has buffered channel of certain capacity
//   - — remain is how many results remain to be read from channel
//   - — ch is how many items are pending in channel
//   - — isError is true if an error was read from a goroutine
//     or SetIsError was invoked
func (g GoResult) String() (s string) {
	if g.goResult == nil {
		return "GoResult_nil"
	}
	return g.goResult.String()
}

// internal GoResultIf
type goResult interface{ GoResultIf }

// GoResult implements GoResultIf
var _ GoResultIf = &GoResult{}

// goResult is internally interface pointer
//   - allows copy of value
//   - points to a channel type wih method-set
type GoResultIf interface {
	// SendError sends error as the final action of a goroutine
	//   - SendError makes a goroutine:
	//   - — awaitable and
	//   - — able to return a fatal error
	//   - — other needs of a goroutine is to initiate and detect cancel and
	//     submit non-fatal errors
	//   - errCh should be a buffered channel large enough for all its goroutines
	//   - — this prevents goroutines from blocking in channel send
	//   - SendError only panics from structural coding problems
	//   - deferrable thread-safe
	SendError(errp *error)
	// ReceiveError is a deferrable function receiving error values from goroutines
	//   - n: number of goroutines to wait for
	//	- n missing: wait for all goroutines.
	//		[NewGoResult]: the number provided to new-function
	//	- [NewGoResult2]: if adds non-zero, wait for adds goroutines.
	//		otherwise wait for number provided to new-function
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
	ReceiveError(errp *error, n ...int) (err error)
	//	- available: the number of results that can be currently collected.
	//		That is len of the result channel, ie.
	//		SendError invocations yet to be collected by ReceiveError
	//	- stillRunning [NewGoResult2] only: the number of created goroutines yet to invoke SendError.
	//		That is cumulative adds less SendError invocations.
	//		If adds is zero, the dimensioned capacity provided to new-function
	//		less SendError invocations
	//   - Thread-safe
	Count() (available, stillRunning int)
	// IsError returns if any goroutine has returned an error
	//	- only for [NewGoResult2]
	IsError() (isError bool)
	// SetIsError sets error state regardless of whether any goroutine has returned an error
	//	- code inspecting IsError will carry out error behavior wihtout the need for a fatal thread exit
	//	- only for [NewGoResult2]
	SetIsError()
	// Remaining returns the number of goroutines that should be awaited
	//	- add: optional add for count-based number of created goroutines
	//	- adds: the cumulative number of add values provided
	//	- — adds allow for not waiting on goroutines that were never created
	//	- if adds is zero, ie. no add was ever provided, adds is the dimensioned
	//		capacity provided to the new-function
	//	- only for [NewGoResult2]
	Remaining(add ...int) (adds int)
	// printable representation
	//	- never panics, never empty string
	String() (s string)
}

// NewGoResult returns the minimum mechanic to make a goroutine awaitable
//   - n is goroutine capacity, default 1
//   - mechanic is buffered channel
//   - a thread-launcher provides a GoResult value of sufficient capacity to its launched threads
//   - exiting threads send an error value that may be nil
//   - the thread-launcher awaits results one by one
//   - to avoid threads blocking prior to exiting, the channel must have sufficient capacity
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
//	  defer g.SendError(&err)
//	  defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)
//
//	  err = …
func NewGoResult(n ...int) (goResult GoResult) { return GoResult{goResult: newGoResultChan(n...)} }

// NewGoResult2 also has [GoResult.IsError] [GoResult.Remaining] [GoResult.SetIsError]
func NewGoResult2(n ...int) (goResult GoResult) {
	return GoResult{goResult: newGoResultStruct(newGoResultChan(n...))}
}
