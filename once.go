/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

// parl.Once is an observable sync.Once with an alternative DoErr method
//   - DoErr invokes a function returning error
//   - IsDone returns whether the Once has been executed
//   - parl.Once is thread-safe and does not require initialization
//   - No thread will return from Once.Do or Once.Doerr until once.Do has completed
type Once struct {
	once   sync.Once
	isDone AtomicBool

	lock   sync.RWMutex
	result InvokeResult
}

// Do calls the function if and only if Do or DoErr is being called for the first time
// for this instance of Once. Thread-safe
func (o *Once) Do(f func()) {
	o.once.Do(func() {
		defer o.isDone.Set()

		f()
	})
}

// DoErr calls the function if and only if Do or DoErr is being called for the first time
// for this instance of Once
//   - didOnce is true if this invocation was the first that did execute f
//   - isPanic is true if f had panic
//   - err is the return value from f or a possible panic
func (o *Once) DoErr(f func() (err error)) (didOnce, isPanic bool, err error) {
	isPanic, err = o.getPreviousResults()
	o.once.Do(func() {
		defer o.isDone.Set()
		defer o.doErr(&isPanic, &err)

		didOnce = true
		isPanic, err = RecoverInvocationPanicErr(f)
	})
	return
}

func (o *Once) IsDone() (isDone bool) {
	return o.isDone.IsTrue()
}

func (o *Once) doErr(isPanicp *bool, errp *error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.result.IsPanic = *isPanicp
	o.result.Err = *errp
}

func (o *Once) getPreviousResults() (isPanic bool, err error) {
	o.lock.RLock()
	defer o.lock.RUnlock()

	isPanic = o.result.IsPanic
	err = o.result.Err

	return
}
