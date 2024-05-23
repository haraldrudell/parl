/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

type AtomicError struct{ err atomic.Pointer[error] }

var _ ErrorSink1 = &AtomicError{}

var _ ErrorSource1 = &AtomicError{}

func (a *AtomicError) AddError(err error) {
	if err == nil {
		return
	}
	for {

		// read current error pointer
		var ep = a.err.Load()
		if ep == nil && a.err.CompareAndSwap(nil, &err) {
			return // stored new error value
		}

		// produce composite error value
		var e2 = perrors.AppendError(*ep, err)
		if a.err.CompareAndSwap(ep, &e2) {
			return // appended to error value
		}
	}
}

func (a *AtomicError) Error() (err error, hasValue bool) {
	if ep := a.err.Load(); ep != nil {
		err = *ep
		hasValue = err != nil
	}
	return
}
