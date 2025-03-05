/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "fmt"

// goResult describes [GoResult] behavior
//   - GoResult makes goroutines awaitable
//   - GoResult implements goResult internally
//   - goResult is package-private interface pointer
//     that allows copy of value
type goResult interface {
	// done executes goroutine exit
	done(err error)
	// ch obtains the error providing channel
	ch() (ch <-chan error)
	// count returns result status
	count() (available, stillRunning int)
	// printable representation
	//	- never panics, never empty string
	//	- “goResult_…”
	fmt.Stringer
}
