/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

const (
	// IncludeClosing is optional argument to IsClosed
	//	- IncludeClosing missing: IsClosed returns true once both Close
	//		was invoked and the stream has been read to end
	//	- IncludeClosing provided: IsClosed returns true if Close was invoked
	IncludeClosing = true
)

// [IncludeClosing] argument to IsClosed methods
type Closing bool
