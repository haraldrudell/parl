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
	lock   sync.Mutex
	ch     chan struct{}
	isDone bool
}

var _ context.Context = &OnceChan{} // OnceChan is a context.Context

func (oc *OnceChan) Done() (ch <-chan struct{}) {
	ch, _ = oc.get(false)
	return
}

func (oc *OnceChan) Cancel() {
	oc.get(true)
}

func (oc *OnceChan) IsDone() (isDone bool) {
	_, isDone = oc.get(false)
	return
}

func (oc *OnceChan) Err() (err error) {
	_, isDone := oc.get(false)
	if isDone {
		err = context.Canceled
	}
	return
}

func (oc *OnceChan) Deadline() (deadline time.Time, ok bool) {
	return
}

func (oc *OnceChan) Value(key any) (value any) {
	return
}

func (oc *OnceChan) get(setCancel bool) (ch chan struct{}, isDone bool) {
	oc.lock.Lock()
	defer oc.lock.Unlock()

	ch = oc.ch
	isDone = oc.isDone
	if ch == nil {
		ch = make(chan struct{})
		oc.ch = ch
	}
	if setCancel && !isDone {
		isDone = true
		oc.isDone = true
		close(ch)
	}
	return
}
