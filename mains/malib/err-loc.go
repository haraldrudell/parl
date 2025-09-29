/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package malib

const (
	// displays error location for errors printed without stack trace
	//	- second argument to [Executable.LongErrors]
	NoLocation ErrLoc = true
)

// [Executable.LongErrors] isLongErrors
type ErrLoc bool
