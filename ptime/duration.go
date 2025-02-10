/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"fmt"
	"math"
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
		// hours and minutes, -24 < d < +24 hours
		var hours = int(d.Hours())
		// minutes 0–59 positive when d negative
		var mins = int(math.Abs(d.Minutes())) % 60
		return strconv.Itoa(hours) + "h" + strconv.Itoa(mins) + "m"
	} else if absValue < 240*time.Hour {
		// -240 < hours < 240
		var hours0 = int(d.Hours())
		// -10 < days < 10
		var days = hours0 / 24
		var hours = hours0 % 24
		if hours < 0 {
			hours = -hours
		}
		return strconv.Itoa(days) + "d" + strconv.Itoa(hours) + "h"
	}
	d = d.Truncate(24 * time.Hour)               // 10+ days
	return strconv.Itoa(int(d.Hours())/24) + "d" // months 36d years 3636d
}

func DurationHMS(d time.Duration) (printableHMS string) {

	// sign
	var sign string
	var dPos time.Duration
	if d < 0 {
		sign = "-"
		dPos = -d
	} else {
		dPos = d
	}

	// hours digits
	var hoursS string
	var hours = uint64(dPos / time.Hour)
	if hours < 10 {
		hoursS = fmt.Sprintf("%02d", hours)
	} else {
		hoursS = fmt.Sprintf("%d", hours)
	}

	var mins = int(dPos / time.Minute % 60)
	var seconds = int(dPos / time.Second % 60)

	printableHMS = fmt.Sprintf("%s%s:%02d:%02d",
		sign,
		hoursS,
		mins,
		seconds,
	)

	return
}
