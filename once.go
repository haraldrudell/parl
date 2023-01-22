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
	once sync.Once

	// temporary

	f    func()
	fErr func() (err error)

	// remnant state

	isDone  AtomicBool
	err     error
	isPanic bool
}

// Do calls the function if and only if Do or DoErr is being called for the first time
// for this instance of Once
func (o *Once) Do(f func()) {
	defer func() {
		o.f = nil
	}()

	o.f = f

	o.once.Do(o.do)
}

// DoErr calls the function if and only if Do or DoErr is being called for the first time
// for this instance of Once
//   - didOnce is true if this invocation was the first that did execute f
//   - isPanic is true if f had panic
//   - err is the return value from f or a possible panic
func (o *Once) DoErr(f func() (err error)) (didOnce, isPanic bool, err error) {
	defer func() {
		isPanic = o.isPanic
		err = o.err
		o.fErr = nil
	}()

	o.fErr = f
	didOnce = !o.isDone.IsTrue()

	o.once.Do(o.doErr)

	return
}

func (o *Once) doErr() {
	var e error
	defer func() {
		o.isDone.Set()
		if o.isPanic = e != nil; o.isPanic {
			o.err = e
		}
	}()

	RecoverInvocationPanic(o.dofErr, &e)
}

func (o *Once) dofErr() {
	o.err = o.fErr()
}

func (o *Once) do() {
	defer o.isDone.Set()

	o.f()
}

func (o *Once) IsDone() (isDone bool) {
	return o.isDone.IsTrue()
}
