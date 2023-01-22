/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pruntime"
)

// ThreadSafeThreadData controls access to a ThreadData object making it thread-safe.
//   - ThreadSafeThreadData does not have initialization
//   - haveThreadID indicates whether data is present
type ThreadSafeThreadData struct {
	haveThreadID parl.AtomicBool

	lock sync.RWMutex
	td   ThreadData
}

// HaveThreadID indicates whether Update has been invoked on this ThreadDataWrap
// object.
func (tw *ThreadSafeThreadData) HaveThreadID() (haveThreadID bool) {
	return tw.haveThreadID.IsTrue()
}

// Update populates the wrapped ThreadData from the stack trace.
func (tw *ThreadSafeThreadData) Update(stack parl.Stack, label string) {
	tw.lock.Lock()
	defer tw.lock.Unlock()

	tw.td.Update(stack, label)
	if stack.ID().IsValid() {
		tw.haveThreadID.Set() // if we know have a vald ThreadID
	}
}

// SetCreator gets preliminary Go identifier: the line invoking Go().
func (tw *ThreadSafeThreadData) SetCreator(cl *pruntime.CodeLocation) {
	tw.lock.Lock()
	defer tw.lock.Unlock()

	tw.td.SetCreator(cl)
}

// Get returns a clone of the wrapped ThreadData object.
func (tw *ThreadSafeThreadData) Get() (thread *ThreadData) {
	tw.lock.RLock()
	defer tw.lock.RUnlock()

	// duplicate ThreadData object
	t := tw.td
	thread = &t

	return
}

// ThreadID returns the thread id of the running thread or zero if
// thread ID is missing.
func (tw *ThreadSafeThreadData) ThreadID() (threadID parl.ThreadID) {
	tw.lock.RLock()
	defer tw.lock.RUnlock()

	return tw.td.ThreadID()
}
