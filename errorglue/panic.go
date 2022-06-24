/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

const panicString = "panic: "

// PanicType is an error originating from a panic
type PanicType struct {
	ErrorChain
}

var _ error = &PanicType{}   // PanicType behaves like an error
var _ Wrapper = &PanicType{} // PanicType has an error chain

func NewPanic(err error) error {
	return &PanicType{*newErrorChain(err)}
}

// Error prepends “panic: ” to the error message
func (w *PanicType) Error() (s string) {
	return panicString + w.ErrorChain.Error()
}
