/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

type AtomicError struct{ err atomic.Pointer[error] }

var _ ErrorSink1 = &AtomicError{}

var _ ErrorSource1 = &AtomicError{}

func (a *AtomicError) AddError(err error) {
	for {
		var e = a.err.Load()
		if e == nil && a.err.CompareAndSwap(nil, &err) {
			return // stored new error value
		}
		var e1 error
		if e != nil {
			e1 = *e
		}
		var e2 = perrors.AppendError(e1, err)
		if e == nil && a.err.CompareAndSwap(e, &e2) {
			return // appended to error value
		}
	}
}

// AddErrorSwap attempts to write newErr if empty or matching oldErr
//   - oldErr: nil or an error returned by [AtomicError.AddErrorSwap] or otherErr
//   - newErr: the error to store
//   - didSwap true: newErr was stored either because of empty or matching oldErr
//   - didSwap false: an error differerent fro oldErr is returned in otherErr
func (a *AtomicError) AddErrorSwap(oldErr, newErr error) (didSwap bool, otherErr error) {

	// if empty, write new value
	var ep = a.err.Load()
	if ep == nil {
		if didSwap = a.err.CompareAndSwap(nil, &newErr); didSwap {
			return // wrote new error return
		}
	}

	// if oldErr, try to write newErr
	if *ep == oldErr {
		if didSwap = a.err.CompareAndSwap(ep, &newErr); didSwap {
			return // updated error return
		}
	}
	otherErr = *ep

	return // didSwap false return: an error exists, and it is not oldErr
}

func (a *AtomicError) Error() (err error, hasValue bool) {
	if ep := a.err.Load(); ep != nil {
		err = *ep
		hasValue = err != nil
	}
	return
}
