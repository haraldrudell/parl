/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

// DeferredErrorSink is a deferrable function that provides
// an error to ErrorSink if
// errp is non-nil pointer to non-nil error
func DeferredErrorSink(errorSink ErrorSink, errp *error) {
	var err error
	if errp == nil {
		return
	} else if err = *errp; err == nil {
		return
	}
	errorSink.AddError(err)
}

// privateErrorSink allows a type with a private addError method to be used as [ErrorSink]
type privateErrorSink struct{ ErrorSink1 }

// NewErrorSinkEndable returns an error sink based on a type with private addError method
func NewErrorSinkEndable(errorSink1 ErrorSink1) (errorSink ErrorSink) {
	return &privateErrorSink{ErrorSink1: errorSink1}
}

// EndErrors is a close-like function noting that AddError will no longer be invoked
//   - if the underlying object does not habe endErrors, EndErrors does nothing
func (p *privateErrorSink) EndErrors() {
	if endable, ok := p.ErrorSink1.(ErrorSink); ok {
		endable.EndErrors()
	}
}

// AddError is a function to submit non-fatal errors
//
// Deprecated: should use [github.com/haraldrudell/parl.ErrorSink]
// possibly the error container [github.com/haraldrudell/parl.ErrSlice]
type AddError func(err error)

// ErrorSink provides send of non-fatal errors
// one at a time
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

// ErrorSink1 provides send of non-fatal errors
// one at a time that cannot be closed
type ErrorSink1 interface {
	// AddError is a function to submit non-fatal errors
	//	- triggers [ErrorSource.WaitCh]
	//	- values are received by [ErrorSource.Error] or [ErrorsSource.Errors]
	AddError(err error)
}

// ErrorSource1 is an error source that is not awaitable
type ErrorSource1 interface {
	// Error returns the next error value
	//	- hasValue true: err is valid
	//	- hasValue false: the error source is currently empty
	Error() (err error, hasValue bool)
}

// ErrorSource provides receive of errors one at a time
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
type Errs interface {
	// ErrorSource provides receive of errors one at a time using
	// WaitCh Error
	ErrorSource
	// Errors returns a slice of errors
	ErrorsSource
}

// ErrorsSource provides receiving multiple
// errors at once
type ErrorsSource interface {
	// Errors returns a slice of errors or nil
	Errors() (errs []error)
}

// absent [parl.AddError] argument
//
// Deprecated: should use [github.com/haraldrudell/parl.ErrorSink]
// possibly the error container [github.com/haraldrudell/parl.ErrSlice]
var NoAddErr AddError
