/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import "time"

// RateRunnerTask describes a rate counter
//   - type local to counter package
type RateRunnerTask interface {
	// Do executes averaging for an accurate timestamp
	Do(at time.Time)
}
