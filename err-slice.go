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
func (e *ErrSlice) Error() (error, bool) { return e.errs.Get() }

// Errors returns a slice of errors or nil
func (e *ErrSlice) Errors() (errs []error) { return e.errs.GetAll() }

// WaitCh waits for the next error, possibly indefinitely
//   - a received channel closes on errors available
//   - the next invocation may return a different channel object
func (e *ErrSlice) WaitCh() (ch AwaitableCh) { return e.errs.DataWaitCh() }

// EndCh awaits the error source closing:
//   - the error source must be read to empty
//   - the error source must be closed by the error-source providing entity
func (e *ErrSlice) EndCh() (ch AwaitableCh) { return e.errs.EmptyCh(CloseAwaiter) }

// AddError is a function to submit non-fatal errors
func (e *ErrSlice) AddError(err error) { e.errs.Send(err) }

// EndCh awaits the error source closing:
//   - the error source must be read to empty
//   - the error source must be closed by the error-source providing entity
func (e *ErrSlice) EndErrors() { e.errs.EmptyCh() }

// AppendErrors collects any errors contained and appends them to errp
func (e *ErrSlice) AppendErrors(errp *error) {
	for _, err := range e.errs.GetAll() {
		*errp = perrors.AppendError(*errp, err)
	}
}
