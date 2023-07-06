/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
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
	hasChannel atomic.Bool // hasChannel provides thread-safe lock-free read of ch
	chLock     sync.Mutex
	// ch is the channel object
	//	- outside the new function, ch is written behind chLock
	ch chan T

	isCloseInvoked atomic.Bool // indicates the channel being closed or about to close
	closeOnce      Once        // [parl.Once] is an observable sync.Once
}

// NewClosableChan returns a channel with idempotent panic-free observable close
//   - ch is an optional non-closed channel object
//   - if ch is not present, an unbuffered channel will be created
//   - cannot use lock in new function
//   - if an unbuffered channel is used, NewClosableChan is not required
func NewClosableChan[T any](ch ...chan T) (cl *ClosableChan[T]) {
	c := ClosableChan[T]{}
	if len(ch) > 0 {
		if c.ch = ch[0]; c.ch != nil {
			c.hasChannel.Store(true)
		}
	}
	return &c
}

// Ch retrieves the channel. Thread-safe
//   - nil is never returned
//   - the channel may already be closed
//   - do not close the channel other than using the Close method
//   - per Go channel close, if one thread is blocked in channel send
//     while another thread closes the channel, a data race occurs
func (c *ClosableChan[T]) Ch() (ch chan T) {
	return c.getCh()
}

// ReceiveCh retrieves the channel. Thread-safe
//   - nil is never returned
//   - the channel may already be closed
//   - do not close the channel other than using the Close method
func (c *ClosableChan[T]) ReceiveCh() (ch <-chan T) {
	return c.getCh()
}

// SendCh retrieves the channel. Thread-safe
//   - nil is never returned
//   - the channel may already be closed
//   - do not close the channel other than using the Close method
//   - per Go channel close, if one thread is blocked in channel send
//     while another thread closes the channel, a data race occurs
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
//   - all invocations have close result in err
//   - didClose indicates whether this invocation closed the channel
//   - if errp is non-nil, it will receive the close result
//   - per Go channel close, if one thread is blocked in channel send
//     while another thread closes the channel, a data race occurs
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
func (cl *ClosableChan[T]) getCh() (ch chan T) {

	// wrap lock in performance-friendly atomic
	//	- by reading hasChannel cl.ch access is thread-safe
	//	- if channel is closed, return whatever ch is
	if cl.hasChannel.Load() || cl.closeOnce.IsDone() {
		return cl.ch
	}

	// ensure a channel is present
	cl.chLock.Lock()
	defer cl.chLock.Unlock()

	if ch = cl.ch; ch == nil {
		ch = make(chan T)
		cl.ch = ch
		cl.hasChannel.Store(true)
	}
	return
}

// doClose is behind [ClosableChan.closeOnce] and
// is therefore only invoked once
func (cl *ClosableChan[T]) doClose() (err error) {

	// ensure a channel exists and close it
	Closer(cl.getCh(), &err)

	return
}
