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
// Duration is more readable than time.Duration.String() that has full ns precision.
func Duration(d time.Duration) (printableDuration string) {
	if d < 10*time.Microsecond {
		return d.String() // ns full precision 314ns
	} else if d < 10*time.Millisecond {
		return d.Round(time.Microsecond).String() // µs 314µs
	} else if d < 1*time.Second {
		return d.Round(time.Millisecond).String() // ms 314ms
	} else if d < 10*time.Second {
		return d.Round(100 * time.Millisecond).String() // 10s 3.1s
	} else if d < 10*time.Minute {
		return d.Round(time.Second).String() // s 31s
	} else if d < 10*time.Hour {
		return d.Round(time.Second).String() // min 5m14s
	} else if d < 24*time.Hour {
		d = d.Round(time.Minute) // h 17h27m
		return strconv.Itoa(int(d.Hours())) + "h" + strconv.Itoa(int(d.Minutes())%60) + "m"
	} else if d < 240*time.Hour {
		d = d.Round(time.Hour) // days 3d15h
		return strconv.Itoa(int(d.Hours())/24) + "d" + strconv.Itoa(int(d.Hours())%24) + "h"
	}
	d = d.Round(24 * time.Hour)
	return strconv.Itoa(int(d.Hours())/24) + "d" // months 36d years 3636d
}
