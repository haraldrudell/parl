/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parltime

import (
	"time"
)

const (
	// UTC time zone
	UTC Tz = iota
	// LOCAL time zone
	LOCAL
)

// Tz indicates what time zone to use
type Tz byte

// OnTimerLocal waits to the next on-the-hour like period
func OnTimerLocal(period time.Duration) (timer *time.Timer) {
	return OnTimer(period, LOCAL)
}

// OnTimerUTC waits to the next on-the-hour like period
func OnTimerUTC(period time.Duration) (timer *time.Timer) {
	return OnTimer(period, UTC)
}

// OnTimer waits to the next on-the-hour like period
func OnTimer(period time.Duration, timeZone Tz) (timer *time.Timer) {
	timer = time.NewTimer(Duro(period, timeZone, time.Now()))
	return
}

// Duro calculates how much time is left of period
func Duro(period time.Duration, timeZone Tz, atTime time.Time) (d time.Duration) {
	d = atTime.Add(period).Truncate(period).Sub(atTime) // value for UTC
	if timeZone == LOCAL {
		_, secondsEastofUTC := atTime.Local().Zone()
		toAdd := -time.Duration(secondsEastofUTC) * time.Second
		d = (d + toAdd) % period // -period < d < period: may be negative
		if d < 0 {
			d += period
		}
	}
	if d == 0 { // 0 <= d < period
		d += period
	}
	return // 0 < d <= period
}
