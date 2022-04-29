/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

/*
ClosableChan wraps channel close.
ClosableChan is an initialization-free channel with a deferable, thread-safe,
idempotent and observable Close method.
Close closes the channel exactly once, recovering panics.
IsClosed provides wether the Close method did execute close.
 var errCh parl.ClosableChan[error]
 go thread(&errCh)
 err, ok := <-errCh.Ch()
 if errCh.isClosed() { // can be inspected
 …
 func thread(errCh *parl.ClosableChan[error]) {
   defer errCh.Close(nil) // will not terminate the process
   errCh.Ch() <- err
*/
type ClosableChan[T any] struct {
	lock sync.Mutex
	ch   chan T // behind lock
	err  error  // behind lock

	closeSelector AtomicBool
	isClosed      AtomicBool
}

// NewClosableChan ensures a chan does not throw
func NewClosableChan[T any](ch ...chan T) (cl *ClosableChan[T]) {
	c := ClosableChan[T]{}
	c.getCh(ch...) // ch... or make provides the channel
	return &c
}

// Ch retrieves the channel
func (cl *ClosableChan[T]) Ch() (ch chan T) {
	return cl.getCh()
}

// Close ensures the channel is closed.
// Close does not panic.
// Close is thread-safe.
// Close does not return until the channel is closed.
// Upon return, all invocations have a possible close error in err.
// if errp is non-nil, it is updated with a possible error
// didClose indicates whether this invocation closed the channel
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
	return cl.isClosed.IsTrue()
}

func (cl *ClosableChan[T]) getCh(ch0 ...chan T) (ch chan T) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if ch = cl.ch; ch == nil {
		if len(ch0) > 0 {
			ch = ch0[0]
		} else {
			ch = make(chan T)
		}
		cl.ch = ch
	}
	return
}

func (cl *ClosableChan[T]) close() (didClose bool, err error) {
	ch := cl.getCh()
	cl.lock.Lock()
	defer cl.lock.Unlock()

	// first thread closes the channel
	if cl.closeSelector.Set() { // this one must be before close
		Closer(ch, &cl.err)
		cl.isClosed.Set() // that one is set after close
		didClose = true
	}
	err = cl.err
	return
}
