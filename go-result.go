/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// GoResult makes any number of goroutines awaitable
//   - number of goroutines must be known at time of new
//   - [NewGoResult] is the simplest, goroutines are awaited by [GoResult.ReceiveError]
//   - [NewGoResult2] also has [IsError] method indicating if any goroutine
//     exited with fatal error
//   - [GoResult.IsValid] true if the GoResult is initialized
//   - [GoResult.SendError](errp *error) deferrable, how goroutine sends results
//   - [GoResult.ReceiveError](errp *error, n ...int) (err error)
//   - [GoResult.Count]() (count int) number of buffered errors
//   - [NewGoResult2] also has:
//   - — [GoResult.IsError]() (isError bool) true if any goroutine returned error
//   - — [GoResult.SetIsError]() sets the error flag manually
//   - — [GoResult.Remaining]() (remaining int) number of goroutines that have yet to exit
//   - —
//   - passed by value
//   - getting around that receiver cannot be interface
//   - receiver is value struct with pointer in the form of an interface
type GoResult struct{ goResult }

func (g GoResult) IsValid() (isValid bool) { return g.goResult != nil }

type goResult interface {
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
	//   - n is number of goroutines to wait for, default 1
	//	- — for [NewGoResult2] default wait for all remaining threads
	//   - errp may be nil
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
	// Count returns number of results that can be currently collected
	//   - Thread-safe
	Count() (count int)
	// IsError returns if any goroutine has returned an error
	//	- only for [NewGoResult2]
	IsError() (isError bool)
	SetIsError()
	// Remaining returns the number of goroutines that have yet to exit
	//	- only for [NewGoResult2]
	Remaining() (remaining int)
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

// NewGoResult2 also has [GoResult.IsError] [GoResult.Remaining]
func NewGoResult2(n ...int) (goResult GoResult) {
	return GoResult{goResult: newGoResultStruct(newGoResultChan(n...))}
}
