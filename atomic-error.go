/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// AtomicError is a thread-safe container for a single error value
//   - [AtomicError.AddError] sets or appends to the error value
//   - [AtomicError.AddErrorSwap] conditionally updates the error value.
//     Used when an existing error value needs to be updated
//   - [AtomicError.Error] returns the current error value and hasValue flag
//   - AtomicError is not closable and holds only one updatable value
//   - AtomicError is not awaitable or readable to empty or iterable
//   - consecutive Error returns the same error value
//   - initialization-free, thread-safe, not awaitable
type AtomicError struct{ err atomic.Pointer[error] }

// AtomicError is an [ErrorSink1] for one error at a time
//   - AtomicError is not closable and holds only one updatable value
var _ ErrorSink1 = &AtomicError{}

// AtomicError is an [ErrorSource1] for one error at a time
//   - AtomicError is not awaitable or readable to empty
//   - consecutive Get returns the same error value
var _ ErrorSource1 = &AtomicError{}

// AddError is a function to submit non-fatal errors
//   - if the container is not empty, err is appended to the current error value
//   - values are received by [ErrorSource1.Error]
func (a *AtomicError) AddError(err error) {

	// try to store new value
	if a.err.Load() == nil && a.err.CompareAndSwap(nil, &err) {
		return // stored new error value
	}
	// err.Load() is not nil

	// append err to error container’s error
	for {
		var ep = a.err.Load()
		var newErrrorValue = perrors.AppendError(*ep, err)
		if a.err.CompareAndSwap(ep, &newErrrorValue) {
			return // appended to error value
		}
	}
}

// AddErrorSwap is thread-safe atomic swap of error values
//   - oldErr: nil or an error returned by [AtomicError.AddErrorSwap] or
//     this methods’s otherErr
//   - newErr: the error to store
//   - didSwap true: newErr was stored either because of empty or matching oldErr
//   - didSwap false: the error held by the container that is differerent from oldErr and
//     is returned in otherErr
//   - AddErrorSwap writes newErr if:
//   - — oldErr is nil and the error container is empty or
//   - — oldErr matches the error held by the error container
func (a *AtomicError) AddErrorSwap(oldErr, newErr error) (didSwap bool, otherErr error) {

	// if empty, write new value
	if a.err.Load() == nil {
		if didSwap = a.err.CompareAndSwap(nil, &newErr); didSwap {
			return // wrote new error return: didSwap true, otherErr nil
		}
	}
	// err.Load() is not nil, didSwap is false

	for {

		// check if error value is oldErr
		var ep = a.err.Load()
		if *ep != oldErr {
			otherErr = *ep
			return // swap fail return: didSwap false, otherErr valid
		}

		// try replacing oldErr
		if didSwap = a.err.CompareAndSwap(ep, &newErr); didSwap {
			return // updated error return: didSwap true, otherErr nil
		}
	}
}

// Error returns the error value
//   - hasValue true: an error was stored
func (a *AtomicError) Error() (err error, hasValue bool) {
	if ep := a.err.Load(); ep != nil {
		err = *ep
		hasValue = true
	}
	return
}
