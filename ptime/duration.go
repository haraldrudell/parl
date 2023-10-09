/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"strconv"
	"time"
)

// Duration formats time.Duration to string with 2 digits precision or better.
//   - Duration is more readable than time.Duration.String() that has full ns precision.
//   - max precision is ns for < 10 µs
//   - min precision is day for >= 10 days
//   - units: ns µs ms s m h d “17h27m”
//   - subsecond digits are rounded with halfway-value rounded away from zero
//   - second and larger values is truncate
func Duration(d time.Duration) (printableDuration string) {
	var absValue time.Duration
	if d >= 0 {
		absValue = d
	} else {
		absValue = -d
	}
	if absValue < 10*time.Microsecond {
		return d.String() // ns full precision 314ns
	} else if absValue < 10*time.Millisecond {
		return d.Round(time.Microsecond).String() // µs 314µs
	} else if absValue < 1*time.Second {
		return d.Round(time.Millisecond).String() // ms 314ms
	} else if absValue < 10*time.Second {
		return d.Round(100 * time.Millisecond).String() // 10s 3.1s
	} else if absValue < 10*time.Minute {
		return d.Truncate(time.Second).String() // s 31s
	} else if absValue < 10*time.Hour {
		return d.Truncate(time.Second).String() // min 5m14s
	} else if absValue < 24*time.Hour {
		d = d.Truncate(time.Minute) // h 17h27m
		return strconv.Itoa(int(d.Hours())) + "h" + strconv.Itoa(int(d.Minutes())%60) + "m"
	} else if absValue < 240*time.Hour {
		d = d.Truncate(time.Hour) // days 3d15h
		return strconv.Itoa(int(d.Hours())/24) + "d" + strconv.Itoa(int(d.Hours())%24) + "h"
	}
	d = d.Truncate(24 * time.Hour)               // 10+ days
	return strconv.Itoa(int(d.Hours())/24) + "d" // months 36d years 3636d
}
