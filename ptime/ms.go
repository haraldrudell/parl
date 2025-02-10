/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import "time"

// Ms gets duration in milliseconds
//   - output is ns precision
//   - 0.123456789 s → “123ms”
//   - 1.2 s → “1.2s”
//   - zero-value → “0s”
func Ms(d time.Duration) string { return d.Truncate(time.Millisecond).String() }
