/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"
)

// parl.Once is an observable sync.Once with an alternative DoErr method
//   - [Once.DoErr] invokes a function returning error recovering a panic
//   - [Once.IsDone] returns whether the Once has been executed, atomic performance
//   - [Once.Result] returns a possible [Once.DoErr] outcome, atomic performance
//   - [Once.Do] is similar to [sync.Once.Do]
//   - parl.Once is thread-safe and does not require initialization
//   - No thread will return from [Once.Do] or [Once.DoErr] until once.Do or once.DoErr has completed
type Once struct {
	// sync.Once is not observable
	once sync.Once
	// isDone indicates if the Once has completed, either by Do or DoErr
	//	- provides observability
	isDone atomic.Bool
	// result is the outcome of a possible DoErr invocation
	//	- if nil, either the Once has not triggered or
	//		it was triggered by Once.Do that does not have a result
	result atomic.Pointer[onceDoErrResult]
}

// onceDoErrResult contains the result of a DoErr invocation
type onceDoErrResult struct {
	isPanic bool
	err     error
}

// DoErr calls the function if and only if Do or DoErr is being called for the first time
// for this instance of Once
//   - didOnce is true if this invocation first and actually invoked doErrFuncArgument
//   - isPanic is true if this or previous invocation did panic
//   - err is either the return value or the panic value from this or previous invocation
//   - thread-safe
//   - —
//   - because sync.Once.Do has fixed signature,
//     Do must be invoke a function wrapper
//   - once.Do must execute for happens before guarantee
//
// Usage:
//
//	var once parl.Once
//	var didTheClose, isPanic, err = once.DoErr(osFile.Close)
func (o *Once) DoErr(doErrFuncArgument func() (err error)) (didOnce, isPanic bool, err error) {

	// wrapper provides the wrapper function for sync.Once.Do
	//	- once.Do must be invoked every time for happens-before
	//	- therefore, wrapper must always be present
	var wrapper = onceDoErr{
		doErrFuncArgument: doErrFuncArgument,
		didOnce:           &didOnce,
		isPanic:           &isPanic,
		errp:              &err,
		Once:              o,
	}

	// execute once.Do to obtain happens-before guarantee
	o.once.Do(wrapper.invokeDoErrFuncArgument)

	if didOnce {
		return // updated by this invocation, isPanic and err are valid return
	}

	// o.once.Do was already triggered and did nothing
	//	- fetch possible previous result
	isPanic, _, err = o.Result()

	return
}

// Do calls the function if and only if Do or DoErr is being called for the first time
// for this instance of Once. Thread-safe
//   - a panic is not recovered
//   - thread-safe
//   - —
//   - once.Do must execute for happens before guarantee
//
// Usage:
//
//	var once parl.Once
//	once.Do(myFunc)
//	…
//	if once.IsDone() …
//	func myFunc() { …
func (o *Once) Do(doFuncArgument func()) {

	// because isDone must be set inside of once.Do,
	// the doFuncArgument argument must be invoked inside a wrapper function
	//	- the wrapper function must have access to doFuncArgument
	//	- because once.Do must be invoked every time,
	//		the wrapper must alway be present
	var d = onceDo{doFuncArgument: doFuncArgument, Once: o}

	o.once.Do(d.invokeF)
}

// IsDone returns true if Once did execute
//   - thread-safe, atomic performance
func (o *Once) IsDone() (isDone bool) { return o.isDone.Load() }

// Result returns the [Once.DoErr] outcome provided with atomic performance
//   - values are only valid if hasResult is true
//   - hasResult is false when:
//   - — the Once has not triggered or
//   - — the Once was triggered by [Once.Do]
//   - thread-safe
func (o *Once) Result() (isPanic bool, hasResult bool, err error) {
	var result = o.result.Load()
	if hasResult = result != nil; !hasResult {
		return // no result available return
	}

	isPanic = result.isPanic
	err = result.err

	return
}
