/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pos

const (
	// stream oiutput to standard error
	Stderr StandardStream = false
	// stream oiutput to standard out
	Stdout StandardStream = true
)

// StandardStream selects a srtandard output stream
//   - [Stderr] [Stdout]
type StandardStream bool
