/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type ErrSlice struct {
	errs AwaitableSlice[error]
}

var _ Errs = &ErrSlice{}
var _ ErrorSink = &ErrSlice{}

func (e *ErrSlice) Error() (error, bool)     { return e.errs.Get1() }
func (e *ErrSlice) Errors() (errs []error)   { return e.errs.GetAll() }
func (e *ErrSlice) WaitCh() (ch AwaitableCh) { return e.errs.DataWaitCh() }
func (e *ErrSlice) AddError(err error)       { e.errs.Send(err) }
