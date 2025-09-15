/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// codeLocation: “runtime/panic.go:914”
//   - err: a main error-chain to traverse for a stack with code location
//   - message: printable string “runtime/panic.go:914”
//   - —
//   - if err or its main error-chain does not have location: empty string
//   - if err or its main error-chain has panic, location is the code line
//     that caused the first panic
//   - if err or its error-chain has location but no panic,
//     location is where the oldest error with stack was created
//   - err is non-nil, no “ at ” prefix
func codeLocation(err error) (message string) {

	// err or err’s error-chain may contain stacks
	//	- any of the stacks may contain a panic
	//	- an error with stack is able to locate any panic it or its chain has
	//	- therefore scan for any error with stack and ask the first one for location
	for e := err; e != nil; e, _, _ = Unwrap(e) {
		if _, ok := e.(ErrorCallStacker); !ok {
			continue // e does not have stack
		}
		var _ = (&errorStack{}).ChainString
		message = e.(ChainStringer).ChainString(ShortSuffix)
		return // found location return
	}

	return // no location return
}
