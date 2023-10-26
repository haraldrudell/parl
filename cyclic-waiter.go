/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// CyclicWait allows any number of threads to wait for a next occurrence.
//   - a parent context may be passed in that on cancel triggers the wait
//     and prevents further cycles
//   - a channel can be obtained that sends one item on the next trig
//     but never closes
//   - a channel can be obtained that closes on next trig
//   - next trig can be awaited
//   - a did-occurer object can be obtained that returns true once the cycle
//     trigs.
//   - a context can be obtained that cancels on the next trig
//   - the cycles can be permanently canceled or trigged and rearmed
type CyclicWait struct {
	parentContext context.Context
	isCancel      atomic.Bool

	lock sync.RWMutex
	ow   OnceWaiter
}

// NewCyclicWait returns a channel that will send one item
// when the context cancels or immediately if the context was already canceled.
func NewCyclicWait(ctx context.Context) (onceReceiver *CyclicWait) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	}
	return &CyclicWait{
		parentContext: ctx,
		ow:            *NewOnceWaiter(ctx),
	}
}

// Ch returns a channel that will emit one item on
// the next trig. It will then not send anything else.
// the channel never closes.
func (cw *CyclicWait) Ch() (ch <-chan struct{}) {
	cw.lock.RLock()
	defer cw.lock.RUnlock()

	return cw.ow.Ch()
}

// Done returns a channel that will close on the next trig or parent context cancel.
// Similar to the Done method of a context.
func (cw *CyclicWait) Done() (done <-chan struct{}) {
	cw.lock.RLock()
	defer cw.lock.RUnlock()

	return cw.ow.Done()
}

// Wait waits until the next trig or parent context cancel.
func (cw *CyclicWait) Wait() {
	done := cw.Done()
	<-done
}

// DidOccurer returns an object with a DidOccur method returning
// true after this cycle has trigged.
func (cw *CyclicWait) DidOccurer() (didOccurer *OnceWaiterRO) {
	cw.lock.RLock()
	defer cw.lock.RUnlock()

	return NewOnceWaiterRO(&cw.ow)
}

// Context returns a context that cancels on the next trig.
func (cw *CyclicWait) Context() (ctx context.Context) {
	cw.lock.RLock()
	defer cw.lock.RUnlock()

	return cw.ow.Context()
}

// Cancel cancels the object and prevents rearming.
func (cw *CyclicWait) Cancel() {
	cw.lock.Lock()
	defer cw.lock.Unlock()

	// trig this cycle
	cw.isCancel.Store(true)
	cw.ow.Cancel()
}

// IsCancel returns whether Cancel has been invoked.
// ISCancel will return false during CancelAndRearm cycles.
func (cw *CyclicWait) IsCancel() (isCancel bool) {
	return cw.isCancel.Load()
}

// CancelAndRearm trigs the object and then rearms unless
// a possible parent context has been canceled.
func (cw *CyclicWait) CancelAndRearm() (wasRearmed bool) {
	cw.lock.Lock()
	defer cw.lock.Unlock()

	// trig this cycle
	cw.ow.Cancel()

	if cw.parentContext.Err() != nil || cw.isCancel.Load() {
		return // ream false: parent context has been canceled
	}

	// rearm: new context
	cw.ow = *NewOnceWaiter(cw.parentContext)
	return
}
