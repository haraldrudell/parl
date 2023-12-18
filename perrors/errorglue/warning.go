/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

const warningString = "warning: "

// warningType is an error with lesser impact
type WarningType struct {
	ErrorChain
}

var _ error = &WarningType{}   // WarningType behaves like an error
var _ Wrapper = &WarningType{} // WarningType has an error chain

func NewWarning(err error) error {
	return &WarningType{*newErrorChain(err)}
}

// Error prepends “Warning: ” to the error message
func (w *WarningType) Error() (s string) {
	return warningString + w.ErrorChain.Error()
}
