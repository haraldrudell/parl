/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"
)

// parl.Once is an observable sync.Once with an alternative DoErr method. Thread-safe
//   - parl.Once is thread-safe and does not require initialization
//   - No thread will return from Once.Do or Once.Doerr until once.Do or once.Doerr has completed
//   - DoErr invokes a function returning error recovering panics
//   - IsDone returns whether the Once has been executed, atomic performance
//   - Result returns DoErr outcome, hasValue indicates if values are present. Atomic eprformance
type Once struct {
	once   sync.Once
	isDone atomic.Bool // isDone indicates if the Once has completed, either by Do or Doerr

	hasResult atomic.Bool // hasResult is true when Once has been completed by DoErr
	// result is the outcome of a possible DoErr invocation
	//	- thread-safe by hasResult atomic
	result invokeResult
}

type invokeResult struct {
	IsPanic bool
	Err     error
}

// Do calls the function if and only if Do or DoErr is being called for the first time
// for this instance of Once. Thread-safe
//   - once.Do must execute for happens before guarantee
func (o *Once) Do(f func()) {
	o.once.Do((&doErrF{f: f, Once: o}).invokeF) // avoid function literal…
}

// DoErr calls the function if and only if Do or DoErr is being called for the first time
// for this instance of Once
//   - didOnce is true if this invocation was the first that did execute f
//   - isPanic is true if f had panic
//   - err is the return value from f or a possible panic
//   - once.Do must execute for happens before guarantee
func (o *Once) DoErr(f func() (err error)) (didOnce, isPanic bool, err error) {

	// avoid function literal here…
	data := doErrData{
		fErr:    f,
		didOnce: &didOnce,
		isPanic: &isPanic,
		errp:    &err,
		Once:    o,
	}
	if o.once.Do(data.invokeFErr); didOnce {
		return // updated by this invocation, isPanic and err are valid return
	}

	// o.once was already triggered but hasResult was false
	//	- isPanic and err may have changed between hasResult and once.Do.
	//	- once.Do is happens before
	isPanic, _, err = o.Result()
	return
}

// IsDone returns whether the Once did execute, provided with atomic performance
func (o *Once) IsDone() (isDone bool) {
	return o.isDone.Load()
}

// Result returns the DoErr outcome provided with atomic performance
//   - only available if hasResult is true
func (o *Once) Result() (isPanic bool, hasResult bool, err error) {
	if hasResult = o.hasResult.Load(); !hasResult {
		return // no result available return
	}

	// provide result protected by atomic
	isPanic = o.result.IsPanic
	err = o.result.Err

	return
}

type doErrData struct {
	fErr    func() (err error)
	didOnce *bool
	isPanic *bool
	errp    *error
	*Once
}

// invokeFErr is behind o.once
func (d *doErrData) invokeFErr() {
	defer d.isDone.Store(true)
	defer d.updateResult()

	*d.didOnce = true
	*d.isPanic, *d.errp = RecoverInvocationPanicErr(d.fErr)
}

func (d *doErrData) updateResult() {
	defer d.hasResult.Store(true)

	d.result.IsPanic = *d.isPanic
	d.result.Err = *d.errp
}

type doErrF struct {
	f func()
	*Once
}

// invokeF is behind o.once
func (d *doErrF) invokeF() {
	defer d.isDone.Store(true)

	d.f()
}
