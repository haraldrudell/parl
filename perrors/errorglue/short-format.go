/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// shortFormat: “message at runtime/panic.go:914”
//   - if err or its error-chain does not have location: “message” like [error.Error]
//   - if err or its error-chain has panic, location is the code line
//     that caused the first panic
//   - if err or its error-chain has location but no panic,
//     location is where the oldest error with stack was created
//   - err is non-nil
func shortFormat(err error) (s string) {

	// append the top frame of the oldest, innermost stack trace code location
	s = codeLocation(err)
	if s != "" {
		s = err.Error() + atStringChain + s
	} else {
		s = err.Error()
	}

	return
}

const (
	// string prepended to code location: “ at ”
	atStringChain = "\x20at\x20"
)
