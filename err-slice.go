/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/perrors"

// ErrSlice is a thread-safe unbound awaitable error container
//   - [ErrSlice.AddError] is a function to submit errors
//   - [ErrSlice.WaitCh] returns a closing channel to await the next error
//   - [ErrSlice.Error] returns the next error value if any
//   - [ErrSlice.Errors] returns a slice of all errors if any
//   - [ErrSlice.EmptyCh] returns a closing channel to await container empty
//     providing deferred-close functionality
//   - ErrSlice features:
//   - — real-time error stream or
//   - — collect errors at end and
//   - — close then read-to-end function
//   - implements [parl.Errs] [parl.ErrorSink]
//
// Usage:
//
//	var errs ErrSlice
//	go fn(&errs)
//	for err := errs.Init(); errs.Condition(&err); {
//	  // process real-time error stream
//	…
//	func fn(errs parl.ErrorSink) {
//	  defer errs.EndErrors()
//	  …
//	  errs.AddError(err)
//
//	var errs ErrSlice
//	fn2(&errs)
//	for _, err := range errs.Errors() {
//	  // process post-invocation errors
//	  …
//	func fn2(errs parl.ErrorSink1) {
//	  errs.AddError(err)
type ErrSlice struct {
	// errs is a thread-safe, unbound awaitable slice of errors
	errs AwaitableSlice[error]
}

// ErrSlice provides error one-at-a-time or all-at-once
var _ Errs = &ErrSlice{}

// ErrSlice has AddError error sink method
var _ ErrorSink = &ErrSlice{}

// Error returns the next error value
//   - hasValue true: err is valid
//   - hasValue false: the error source is empty
func (e *ErrSlice) Error() (err error, hasValue bool) { return e.errs.Get() }

// Errors returns a slice of errors or nil
func (e *ErrSlice) Errors() (errs []error) { return e.errs.GetAll() }

// WaitCh waits for the next error, possibly indefinitely
//   - a received channel that closes whenever errors becoming available
//   - subsequent invocations may return different channel values
func (e *ErrSlice) WaitCh() (ch AwaitableCh) { return e.errs.DataWaitCh() }

// EndCh awaits the error source closing:
//   - the error source must be read to empty
//   - the error source must be closed by the error-source providing entity
func (e *ErrSlice) EndCh() (ch AwaitableCh) { return e.errs.EmptyCh() }

// AddError is a function to submit non-fatal errors
func (e *ErrSlice) AddError(err error) { e.errs.Send(err) }

// EndCh awaits the error source closing:
//   - the error source must be read to empty
//   - the error source must be closed by the error-source providing entity
func (e *ErrSlice) EndErrors() { e.errs.Close() }

// AppendErrors collects any errors contained in ErrSlice and
// appends them to errp
//   - errp: where errors are aggregated to
func (e *ErrSlice) AppendErrors(errp *error) {
	for _, err := range e.errs.GetAll() {
		*errp = perrors.AppendError(*errp, err)
	}
}

// Seq allows for ErrSlice to be used in a for range clause
//   - each value is provided to yield
//   - iterates until yield retuns false or
//   - the slice was empty and in drain-close states
//   - thread-safe
//
// Usage:
//
//	for value := range errSlice.Seq {
//	  value…
//	}
//	// the AwaitableSlice was empty and in drain-closed state
func (e *ErrSlice) Seq(yield func(value error) (keepGoing bool)) { e.errs.Seq(yield) }
