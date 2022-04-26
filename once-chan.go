/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync"
	"time"
)

/*
OnceChan is a semaphore implementing the Context with Cancel interface.
Whenever a context is required for cancellation, a OnceChan can be used
in its place.
Unlike context, OnceChan requires no initialization.
Similar to Context, OnceChan can be waited on like a channel using Done().
OnceChan can be inspected using IsDone() or Err().
OnceChan is cancelled using .Cancel()

 var semaphore OnceChan
 go func() {
   <-onceChan.Done()
 }()
 …
 semaphore.Cancel()
 …
 semaphore.IsDone()
*/
type OnceChan struct {
	initOnce  sync.Once
	closeOnce sync.Once
	ch        chan struct{}
	isDone    AtomicBool
}

var _ context.Context = &OnceChan{} // OnceChan is a context.Context
var _ CancelContext = &OnceChan{}   // OnceChan is a parl.CancelContext

// Done provides the channel to wait for
func (oc *OnceChan) Done() (ch <-chan struct{}) {
	return oc.get()
}

func (oc *OnceChan) Cancel() {
	ch := oc.get() // ensure we have a channel
	oc.closeOnce.Do(func() {
		var err error
		if Closer(ch, &err); err != nil {
			Infallible(err)
		}
		oc.isDone.Set()
	})
}

func (oc *OnceChan) IsDone() (isDone bool) {
	return oc.isDone.IsTrue()
}

// Err is valid for cancel
func (oc *OnceChan) Err() (err error) {
	if oc.isDone.IsTrue() {
		err = context.Canceled
	}
	return
}

// Deadline is not implemented
func (oc *OnceChan) Deadline() (deadline time.Time, ok bool) {
	return
}

// Value is not implemented
func (oc *OnceChan) Value(key any) (value any) {
	return
}

// get facilitates initialization-free
func (oc *OnceChan) get() (ch chan struct{}) {

	// ensure we have a channel
	oc.initOnce.Do(func() {
		if oc.ch == nil {
			oc.ch = make(chan struct{})
		}
	})

	return oc.ch
}
