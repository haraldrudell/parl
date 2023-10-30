/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
)

// ClosableChan wraps a channel with thread-safe idempotent panic-free observable close.
//   - ClosableChan is initialization-free
//   - Close is deferrable
//   - IsClosed provides wether the channel is closed
//
// Usage:
//
//	var errCh parl.ClosableChan[error]
//	go thread(&errCh)
//	err, ok := <-errCh.Ch()
//	if errCh.IsClosed() { // can be inspected
//	…
//
//	func thread(errCh *parl.ClosableChan[error]) {
//	  var err error
//	  …
//	  defer errCh.Close(&err) // will not terminate the process
//	  errCh.Ch() <- err
type ClosableChan[T any] struct {
	// ch0 is the channel object
	//	- ability to initialize ch0 in the constructor
	//	- ability to update ch0 after creation
	//	- ch0 therefore must be pointer
	//	- ch0 must offer thread-safe access and update

	// ch0 as provided by contructor or nil
	ch0 chan T
	// ch0 provided post-constructor because ch0 nil
	chp atomic.Pointer[chan T]

	// indicates the channel about to close or closed
	//	- because the channel may transfer data, it cannot be inspected for being closed
	isCloseInvoked atomic.Bool
	// [parl.Once] is an observable sync.Once
	//	- indicates that the channel is closed
	closeOnce Once
}

// NewClosableChan returns a channel with idempotent panic-free observable close
//   - ch is an optional non-closed channel object
//   - if ch is not present, an unbuffered channel will be created
//   - cannot use lock in new function
//   - if an unbuffered channel is used, NewClosableChan is not required
func NewClosableChan[T any](ch ...chan T) (closable *ClosableChan[T]) {
	var ch0 chan T
	if len(ch) > 0 {
		ch0 = ch[0] // if ch is present, apply it
	}
	return &ClosableChan[T]{ch0: ch0}
}

// Ch retrieves the channel as bi-directional. Thread-safe
//   - nil is never returned
//   - the channel may be closed, use IsClosed to determine
//   - do not close the channel other than using Close method
//   - per Go channel close, if one thread is blocked in channel send
//     while another thread closes the channel, a data race occurs
//   - thread-safe solution is to set an additional indicator of
//     close requested and then reading the channel which
//     releases the sending thread
func (c *ClosableChan[T]) Ch() (ch chan T) {
	return c.getCh()
}

// ReceiveCh retrieves the channel as receive-only. Thread-safe
//   - nil is never returned
//   - the channel may already be closed
func (c *ClosableChan[T]) ReceiveCh() (ch <-chan T) {
	return c.getCh()
}

// SendCh retrieves the channel as send-only. Thread-safe
//   - nil is never returned
//   - the channel may already be closed
//   - do not close the channel other than using the Close method
//   - per Go channel close, if one thread is blocked in channel send
//     while another thread closes the channel, a data race occurs
//   - thread-safe solution is to set an additional indicator of
//     close requested and then reading the channel which
//     releases the sending thread
func (c *ClosableChan[T]) SendCh() (ch chan<- T) {
	return c.getCh()
}

// IsClosed indicates whether the channel is closed. Thread-safe
//   - includePending: because there is a small amount of time between
//   - — a thread discovering the channel closed and
//   - — closeOnce indicating close complete
//   - includePending true includes a check for the channel being about
//     to close
func (c *ClosableChan[T]) IsClosed(includePending ...bool) (isClosed bool) {
	if len(includePending) > 0 && includePending[0] {
		return c.isCloseInvoked.Load()
	}
	return c.closeOnce.IsDone()
}

// Close ensures the channel is closed
//   - Close does not return until the channel is closed.
//   - thread-safe panic-free deferrable observable
//   - all invocations have the same close result in err
//   - didClose indicates whether this invocation closed the channel
//   - if errp is non-nil, it will receive the close result
//   - per Go channel close, if one thread is blocked in channel send
//     while another thread closes the channel, a data race occurs
//   - thread-safe, panic-free, deferrable, idempotent
func (cl *ClosableChan[T]) Close(errp ...*error) (didClose bool, err error) {

	// ensure isCloseInvoked true: channel is about to close
	cl.isCloseInvoked.CompareAndSwap(false, true)

	// hasResult indicates that close did already complete
	// and err was obtained with atomic performance
	var hasResult bool
	_, hasResult, err = cl.closeOnce.Result()

	// first invocation closes the channel
	//	- subsequent invocations await close complete
	//		and return the close result
	if !hasResult {
		didClose, _, err = cl.closeOnce.DoErr(cl.doClose)
	}

	// update errp if present
	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil {
			*errp0 = err
		}
	}

	return
}

// getCh gets or initializes the channel object [ClosableChan.ch]
func (c *ClosableChan[T]) getCh() (ch chan T) {
	if ch = c.ch0; ch != nil {
		return // channel from constructor return
	}
	for {
		if chp := c.chp.Load(); chp != nil {
			ch = *chp
			return // chp was present return
		}
		if ch == nil {
			ch = make(chan T)
		}
		if c.chp.CompareAndSwap(nil, &ch) {
			return // chp updated return
		}
	}
}

// doClose is behind [ClosableChan.closeOnce] and
// is therefore only invoked once
//   - separate function because provided to Once
func (cl *ClosableChan[T]) doClose() (err error) {

	// ensure a channel exists and close it
	Closer(cl.getCh(), &err)

	return
}
