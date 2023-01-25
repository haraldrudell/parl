/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// ErrorString is similar to the private type errors.errorString
type ErrorString struct {
	s string
}

// New returns an error that formats as the given text.
// Each call to New returns a distinct error value even if the text is identical.
func New(text string) error {
	return &ErrorString{text}
}

func (e *ErrorString) Error() string {
	return e.s
}
