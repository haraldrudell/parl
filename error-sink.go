/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"iter"

	"github.com/haraldrudell/parl/perrors"
)

var (
	// NoErrorSink1 is value to use when no [ErrorSink1] is provided
	NoErrorSink1 ErrorSink1
)

// ErrorSink receives non-fatal errors
//   - closable version of [ErrorSink1]
//   - addresses threads’ need to submit non-fatal errors
//   - closability allows for iteration and makes awaitable
//   - [ErrSlice] is thread-safe, multiple-errors, closable implementation
//   - [DeferredErrorSink] is a deferred function collecting to ErrorSink1
//   - error retrieval: [Errs] [ErrorSource] [ErrorSource1] [ErrsIter] [ErrorsSource]
//   - [NewErrorSinkEndable] wraps ErrorSink1 to be usable as ErrorSink
type ErrorSink interface {
	// AddError is a function to submit non-fatal errors
	//	- triggers [ErrorSource.WaitCh]
	//	- values are received by [ErrorSource.Error] or [ErrorsSource.Errors]
	AddError(err error)
	// EndErrors optionally indicates that no more AddError
	// invocations will occur
	//	- enables triggering of [ErrorSource.EndCh]
	EndErrors()
}

// ErrorSink1 receives non-fatal errors
//   - the receiver is not guaranteed to be closable
//   - addresses threads’ need to submit non-fatal errors
//   - [ErrorSink] is closable interface
//   - [ErrSlice] is thread-safe, multiple-errors, closable implementation
//   - [AtomicError] is thread-safe, single-value non-close implementation
//   - [DeferredErrorSink] is a deferred function collecting to ErrorSink1
//   - error retrieval: [Errs] [ErrorSource] [ErrorSource1] [ErrsIter] [ErrorsSource]
type ErrorSink1 interface {
	// AddError submits a non-fatal error
	//	- err may or may not be allowed to be nil
	AddError(err error)
}

// ErrorSource1 is an error source that is not awaitable
//   - implemented by [ErrSlice]
//   - implemented by [AtomicError]
//   - [ErrorSource] is awaitable interface
//   - [Errs] interface can also provide slice
//   - [ErrsIter] is iterable and awaitable
//   - [DeferredErrorSource] collects errors from ErrorSource1
//   - [ErrorSink1] provides error values
type ErrorSource1 interface {
	// Error returns the next error value
	//	- hasValue true: err is valid
	//	- hasValue false: the error source is currently empty
	Error() (err error, hasValue bool)
}

// ErrorSource provides receive of errors one at a time
//   - implemented by [ErrSlice]
//   - [ErrorSource1] interface is not awaitable
//   - [Errs] interface can also provide slice
//   - [ErrsIter] is iterable and awaitable
//   - [DeferredErrorSource] collects errors from ErrorSource1
//   - [ErrorSink] provides error values
type ErrorSource interface {
	ErrorSource1
	// WaitCh waits for the next error, possibly indefinitely
	//	- each invocation returns a channel that closes on errors available
	//	- — invocations may return different channel values
	//	- the next invocation may return a different channel object
	WaitCh() (ch AwaitableCh)
	// EndCh awaits the error source closing:
	//	- the error source must be read to empty
	//	- the error source must be closed by the error-source providing entity
	EndCh() (ch AwaitableCh)
}

// Errs provides receiving errors,
// one at a time or multiple
//   - implemented by [ErrSlice]
//   - [ErrorSource1] interface is not awaitable
//   - [ErrsIter] is iterable and awaitable
//   - [DeferredErrorSource] collects errors from ErrorSource1
//   - [ErrorSink] provides error values
type Errs interface {
	// ErrorSource provides receive of errors one at a time using
	// WaitCh Error
	ErrorSource
	// Errors returns a slice of errors
	ErrorsSource
}

// ErrsIter provides an error source that:
//   - is iterable, awaitable and closable
//   - can return errors one-at-a-time, in iteration or all-at-once
//   - implemented by [ErrSlice]
//   - [ErrorSource1] interface is not awaitable
//   - [DeferredErrorSource] collects errors from ErrorSource1
type ErrsIter interface {
	// Error WaitCh EndCh Errors
	Errs
	// Seq is an iterator over sequences of individual errors.
	// When called as seq(yield), seq calls yield(v) for each value v in the sequence,
	// stopping early if yield returns false.
	Seq(yield func(err error) (keepGoing bool))
}

// ErrsIter.Seq is iters.Seq[error]
var _ = func(e ErrsIter) (seq iter.Seq[error]) { return e.Seq }

// ErrorsSource provides receiving multiple
// errors at once
//   - implemented by [ErrSlice]
//   - typically ussed via [Errs] or [ErrsIter]
type ErrorsSource interface {
	// Errors returns a slice of errors or nil
	Errors() (errs []error)
}

// DeferredErrorSink adds a possible error in errp to errorSink
//   - errorSink: any type of error sink: atomic or endable
//   - errp: may be nil
//   - deferrable, thread-safe if errorSink is thread-safe
func DeferredErrorSink(errorSink ErrorSink1, errp *error) {
	var err error
	if errp == nil {
		return
	} else if err = *errp; err == nil {
		return
	}
	errorSink.AddError(err)
}

// DeferredErrorSource is a deferrable function that appends
// all errors in errorSource to errp
//   - empties an error source but does not wait for it to close
//   - if errorSource is thread-safem it may be shared by multiple threads
func DeferredErrorSource(errorSource ErrorSource1, errp *error) {

	// errorSink may be [AtomicError] that does not empty

	if errs, ok := errorSource.(Errs); ok {
		// eSource is an error source where each Error removes
		// from the source
		for {
			var e, hasError = errs.Error()
			if !hasError {
				return
			}
			*errp = perrors.AppendError(*errp, e)
		}
	}

	// single-error source like [parl.AtomicError]
	var e, hasError = errorSource.Error()
	if !hasError {
		return
	}
	*errp = perrors.AppendError(*errp, e)
}

// privateErrorSink allows a type with a private addError method to be used as [ErrorSink]
type privateErrorSink struct{ ErrorSink1 }

// NewErrorSinkEndable wraps an [ErrorSink1] to be a closable and awaitable [ErrorSink]
//   - if the implementation does not support EndErrors, EndErrors does nothing
func NewErrorSinkEndable(errorSink1 ErrorSink1) (errorSink ErrorSink) {
	return &privateErrorSink{ErrorSink1: errorSink1}
}

// EndErrors is a close-like function noting that AddError will no longer be invoked
//   - if the underlying object does not have endErrors, EndErrors does nothing
func (p *privateErrorSink) EndErrors() {
	if endable, ok := p.ErrorSink1.(ErrorSink); ok {
		endable.EndErrors()
	}
}
