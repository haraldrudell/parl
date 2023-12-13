/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/perrors"

const (
	// minimum error channel length
	minGoResultLength = 1
)

// GoResult awaits thread outcomes using the minimum to make a goroutine awaitable
type GoResult chan error

// NewGoResult returns the minimum mechanic to make a goroutine awaitable
//   - mechanic is buffered channel
//   - a thread-launcher provides a GoResult value of sufficient capacity to its launched threads
//   - exiting threads send an error value that may be nil
//   - the thread-launcher awaits results one by one
//   - to avoid threads blocking prior to exiting, the channel must have sufficient capacity
//   - n is goroutine capacity, default 1
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
func NewGoResult(n ...int) (goResult GoResult) {
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < minGoResultLength {
		n0 = minGoResultLength
	}
	return make(chan error, n0)
}

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
func (g GoResult) SendError(errp *error) {
	if errp == nil {
		panic(NilError("errp"))
	}
	didSend, isNilChannel, isClosedChannel, err := ChannelSend(g, *errp, SendNonBlocking)
	if didSend {
		return // error value sent return
	} else if isNilChannel {
		err = perrors.ErrorfPF("fatal: error channel nil: %w", err)
	} else if isClosedChannel {
		err = perrors.ErrorfPF("fatal: error channel closed: %w", err)
	} else if err != nil {
		err = perrors.ErrorfPF("fatal: panic when sending on error channel: %w", err)
	} else {
		err = perrors.NewPF("fatal: error channel blocking on send")
	}
	panic(err)
}

// ReceiveError is a deferrable function receiving error values from goroutines
//   - n is number of goroutines to wait for, default 1
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
func (g GoResult) ReceiveError(errp *error, n ...int) (err error) {
	if g == nil {
		panic(NilError("GoResult"))
	}
	var n0 int
	if len(n) > 0 {
		n0 = n[0]
	}
	if n0 < minGoResultLength {
		n0 = minGoResultLength
	}

	// await n0 goroutine results
	for ; n0 > 0; n0-- {

		// blocks here
		//	- wait for a result from a goroutine
		var e = <-g
		if e == nil {
			continue // good return: ignore
		}
		e = perrors.Stack(e)
		err = perrors.AppendError(err, e)
	}
	if errp != nil {
		*errp = perrors.AppendError(*errp, err)
	}

	return
}

// Count returns number of results that can be collected. Thread-safe
func (g GoResult) Count() (count int) { return len(g) }
