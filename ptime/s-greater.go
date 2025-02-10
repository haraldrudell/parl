/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"
)

// SGreater compares two times after rounding them to second precision
func SGreater(t1 time.Time, t2 time.Time) bool {
	return t1.Truncate(time.Second).After(t2.Truncate(time.Second))
}
