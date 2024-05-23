/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package watchfs

import "github.com/haraldrudell/parl"

type eSink interface{ addError(err error) }
type eSinkEnd interface{ endErrors() }
type e struct{ privateErrorSink eSink }

var _ parl.ErrorSink = &e{}

// newErrorSink returns a [parl.ErrorSink] based on package-private methods
//   - privateErrorSink must implement addError(err error)
//   - privateErrorSink may implement endErrors()
func newErrorSink(privateErrorSink eSink) (errorSink parl.ErrorSink) {
	return parl.NewErrorSinkEndable(&e{privateErrorSink: privateErrorSink})
}

func (e *e) AddError(err error) { e.privateErrorSink.addError(err) }
func (e *e) EndErrors() {
	if errorSink, ok := e.privateErrorSink.(eSinkEnd); ok {
		errorSink.endErrors()
	}
}
