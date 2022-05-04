/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import "github.com/haraldrudell/parl"

// waiter allows to use any of observable parl.WaitGroup or parl.TraceGroup
type waiter interface {
	parl.Waiter
}
