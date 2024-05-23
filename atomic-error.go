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

func (a *AtomicError) Error() (err error, hasValue bool) {
	if ep := a.err.Load(); ep != nil {
		err = *ep
		hasValue = err != nil
	}
	return
}
