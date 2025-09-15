/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// warningType is an error with lesser impact
type WarningType struct {
	ErrorChain
}

// WarningType behaves like an error
var _ error = &WarningType{}

// WarningType has an error chain
var _ Unwrapper = &WarningType{}

func NewWarning(err error) (warning error) {
	return &WarningType{*newErrorChain(err)}
}

// Error prepends “Warning: ” to the error message
func (w *WarningType) Error() (s string) {
	return warningString + w.ErrorChain.Error()
}

const (
	warningString = "warning: "
)
