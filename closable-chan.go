/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

/*
ClosableChan holds a channel.
ClosableChan has a thread-safe re-entrant Close method
closing the channel exactly once without panics.
IsClosed indicates wether the channel has been closed
*/
type ClosableChan[T any] struct {
	ch       chan T
	err      error
	isClosed AtomicBool
	once     sync.Once
}

// NewCloser ensures a chan does not throw
func NewCloser[T any](ch chan T) (cl *ClosableChan[T]) {
	return &ClosableChan[T]{ch: ch}
}

// Ch retrieves the channel
func (cl *ClosableChan[T]) Ch() (ch chan T) {
	return cl.ch
}

// Close ensures the channel is closed.
// Close does not panic.
// Close is thread-safe.
// Close does not return until the channel is closed.
// Upon return, all invocations have a possible close error in err.
// if errp is non-nil, it is updated with error status
func (cl *ClosableChan[T]) Close(errp ...*error) (err error, didClose bool) {

	// first thread closes the channel
	cl.once.Do(func() {
		defer Recover(Annotation(), &cl.err, nil)

		didClose = true
		cl.isClosed.Set()
		close(cl.ch)
	})

	// all threads provide the result
	err = cl.err
	if len(errp) > 0 {
		*errp[0] = err
	}
	return
}

// IsClosed indicates whether the Close method has been invoked
func (cl *ClosableChan[T]) IsClosed() (isClosed bool) {
	return cl.isClosed.IsTrue()
}

// Closer closes a channels and can be deferred
// and does not panic
func Closer[T any](ch chan T, errp *error) {
	defer Recover(Annotation(), errp, nil)

	close(ch)
}
