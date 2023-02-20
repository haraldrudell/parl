/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

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
//	if errCh.isClosed() { // can be inspected
//	…
//
//	func thread(errCh *parl.ClosableChan[error]) {
//	  var err error
//	  …
//	  defer errCh.Close(&err) // will not terminate the process
//	  errCh.Ch() <- err
type ClosableChan[T any] struct {
	hasChannel AtomicBool
	chLock     sync.Mutex
	ch         chan T // behind lock

	closeOnce Once
}

// NewClosableChan returns a channel with thread-safe idempotent panic-free observable close
func NewClosableChan[T any](ch ...chan T) (cl *ClosableChan[T]) {
	c := ClosableChan[T]{}
	c.getCh(ch...) // ch... or make provides the channel
	return &c
}

// Ch retrieves the channel
func (cl *ClosableChan[T]) Ch() (ch chan T) {
	return cl.getCh()
}

// Close ensures the channel is closed
//   - Close does not return until the channel is closed.
//   - panic-free thread-safe deferrable observable
//   - all invocations have close result in err
//   - didClose indicates whether this invocation closed the channel
func (cl *ClosableChan[T]) Close(errp ...*error) (didClose bool, err error) {

	// first thread closes the channel
	// all threads provide the result
	didClose, err = cl.close()

	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil {
			*errp0 = err
		}
	}

	return
}

// IsClosed indicates whether the Close method has been invoked
func (cl *ClosableChan[T]) IsClosed() (isClosed bool) {
	return cl.closeOnce.IsDone()
}

func (cl *ClosableChan[T]) getCh(ch0 ...chan T) (ch chan T) {

	// wrap lock in performance-friendly atomic
	if cl.hasChannel.IsTrue() {
		return cl.ch
	}

	// ensure a channel is present
	cl.chLock.Lock()
	defer cl.chLock.Unlock()

	if cl.closeOnce.IsDone() {
		return // already closed return
	}

	if ch = cl.ch; ch == nil {
		if len(ch0) > 0 {
			ch = ch0[0]
		} else {
			ch = make(chan T)
		}
		cl.ch = ch
	}
	cl.hasChannel.Set()
	return
}

func (cl *ClosableChan[T]) close() (didClose bool, err error) {

	// provide result with atomic performance
	var hasResult bool
	if _, hasResult, err = cl.closeOnce.Result(); hasResult {
		return // already closed return
	}

	didClose, _, err = cl.closeOnce.DoErr(cl.doClose)
	return
}

func (cl *ClosableChan[T]) doClose() (err error) {
	cl.chLock.Lock()
	defer cl.chLock.Unlock()

	if cl.hasChannel.IsFalse() {
		return // no channel to close return
	}

	Closer(cl.ch, &err)
	return
}
