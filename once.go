/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync"

// parl.Once is an observable sync.Once with an alternative DoErr method. Thread-safe
//   - parl.Once is thread-safe and does not require initialization
//   - No thread will return from Once.Do or Once.Doerr until once.Do or once.Doerr has completed
//   - DoErr invokes a function returning error recovering panics
//   - IsDone returns whether the Once has been executed, atomic performance
//   - Result returns DoErr outcome, hasValue indicates if values are present. Atomic eprformance
type Once struct {
	once   sync.Once
	isDone AtomicBool // isDone indicates if the Once has completed, either by Do or Doerr

	hasResult AtomicBool   // hasResult is true when Once was completed by DoErr
	result    InvokeResult // result is the outcome of a possible DoErr invocation
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
	isPanic, _, err = o.Result()
	o.once.Do(func() {
		defer o.isDone.Set()
		defer o.doErr(&isPanic, &err)

		didOnce = true
		isPanic, err = RecoverInvocationPanicErr(f)
	})
	return
}

// IsDone returns whether the Once did execute, provided with atomic performance
func (o *Once) IsDone() (isDone bool) {
	return o.isDone.IsTrue()
}

// Result returns the DoErr outcome provided with atomic performance
//   - only available if hasResult is true
func (o *Once) Result() (isPanic bool, hasResult bool, err error) {
	if hasResult = o.hasResult.IsTrue(); !hasResult {
		return // no result available return
	}

	// provide result protected by atomic
	isPanic = o.result.IsPanic
	err = o.result.Err

	return
}

func (o *Once) doErr(isPanicp *bool, errp *error) {
	o.result.IsPanic = *isPanicp
	o.result.Err = *errp
	o.hasResult.Set()
}
