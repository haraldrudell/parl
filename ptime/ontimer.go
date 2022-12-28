/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"

	"github.com/haraldrudell/parl/perrors"
)

// OnTimer returns a time.Timer that waits from t to the next period-multiple from zero time.
//   - period must be greater than zero or panic
//   - t contains time zone that matters for durations 24 h or longer, typicaly this is time.Local
//   - default t is time.Now()
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

// Duro returns the duration in nanoseconds until the next duration-multiple from zero time.
//   - The period starts from atTime
//   - time zone for multiple-calculation is defined by atTime, often time.Local
//   - time zone matters for 24 h or longer durations
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
