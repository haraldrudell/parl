/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"

	"github.com/haraldrudell/parl/perrors"
)

// OnTimer waits from now to the next on-the-hour-like period in local time zone
func OnTimer(period time.Duration, t ...time.Time) (timer *time.Timer) {
	var t0 time.Time
	if len(t) > 0 {
		t0 = t[0]
	}
	if t0.IsZero() {
		t0 = time.Now()
	}
	return time.NewTimer(Duro(period, t0))
}

// Duro returns the duration to the next triggering time.
// period is the period duration.
// atTime is the time from which to calculate duration.
// atTime location contains the time zone for calculations of day and longer durations.
func Duro(period time.Duration, atTime time.Time) (d time.Duration) {
	if period <= 0 {
		panic(perrors.Errorf("Duro with non-positive period: %s", period))
	}
	d = atTime.Add(period).Truncate(period).Sub(atTime) // value for UTC: 0 <= d < period
	_, secondsEastofUTC := atTime.Zone()
	toAdd := -time.Duration(secondsEastofUTC) * time.Second
	d = (d + toAdd) % period // -period < d < period: may be negative
	if d < 0 {
		d += period // 0 <= d < period
	}
	if d == 0 {
		d += period // 0 < d <= period
	}
	return // 0 < d <= period
}
