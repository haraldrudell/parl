/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// AddError is a function to submit non-fatal errors
//
// Deprecated: should use [github.com/haraldrudell/parl.ErrorSink]
// possibly the error container [github.com/haraldrudell/parl.ErrSlice]
type AddError func(err error)

// ErrorSink provides send of non-fatal errors
// one at a time
type ErrorSink interface {
	// AddError is a function to submit non-fatal errors
	AddError(err error)
}

// ErrorSource provides receive of errors one at a time
type ErrorSource interface {
	// WaitCh waits for the next error, possibly indefinitely
	//	- a received channel closes on errors available
	//	- the next invocation may return a different channel object
	WaitCh() (ch AwaitableCh)
	// Error returns the next error value
	//	- hasValue true: err is valid
	//	- hasValue false: the error source is empty
	Error() (err error, hasValue bool)
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
