/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"math"
	"time"
)

// Epoch represents a time value in 64-bit integral type that
// can be stored in atomic integer like [Atomic64[Epoch]]
//   - allowable year range is 1678–2261 inclusive
//   - 250308 defined as Unix epoch with nanosecond precision:
//     nanoseconds elapsed since January 1, 1970 UTC
//   - time package does not have a defined type for Unixnano but uses int64
//   - epoch is internally modified Unix Epoch with nanosecond precision
//   - [time.Now] on macOS is μs precision
type Epoch int64

// EpochZeroValue is the zero-value for Epoch
//   - the value of an unitialized Epoch field
//   - corresponds to January 1, 1970 UTC
var EpochZeroValue Epoch

// EpochZeroValueTime is the [Epoch.Time] for Epoch zero-value
//   - the Time value returned by an unitialized Epoch field
//   - corresponds to January 1, 1970 UTC
var EpochZeroValueTime = EpochZeroValue.Time()

// EpochNow translates a time value to a 64-bit integral value that can be
// stored with atomic access in [Atomic64[Epoch]]
//   - t present: a time in [time.Local] or [time.UTC] that can be zero-value
//   - — allowable year range is 1678–2261 inclusive
//   - — zero-time is allowed
//   - t missing: now is used
//   - — [time.Now] on macOS is μs precision
//   - —
//   - [Epoch.Time] converts back
//   - epoch is stable across executable invocations
func EpochNow(t ...time.Time) (epoch Epoch) {

	// get t0
	var t0 time.Time
	if len(t) > 0 {
		t0 = t[0]
	} else {
		t0 = time.Now()
	}

	if t0.IsZero() {
		epoch = epochZeroTime
		return // zero value return
	}

	if t0.Before(minTime) || t0.After(maxTime) {
		epoch = epochInvalid
		return
	}

	epoch = Epoch(t0.UnixNano())
	return
}

// Time returns [time.Time] in [time.Local] corresponding to epoch
//   - time is in [time.Local], may be zero Time
//   - invalid epoch returns zero Time
//   - — [Epoch.IsValid] determines if epoch is valid
//   - —
//   - epoch is stable across executable invocations
func (epoch Epoch) Time() (t time.Time) {
	if epoch == epochZeroTime || epoch == epochInvalid {
		return // epoch zero means time.Time{} ie. time.IsZero()
	}

	var nsec = int64(epoch) % nanoPerSecond
	var sec = int64(epoch) / nanoPerSecond
	t = time.Unix(sec, nsec)
	return
}

// IsValid returns true if epoch is not zero-time, ie. Epoch(0) corredsponding to time.TIME{} and Time.IsZero
func (epoch Epoch) IsValid() (isValid bool) {
	if epoch == epochZeroTime {
		isValid = true
		return
	} else if epoch == epochInvalid {
		return
	}
	// get corresponding time value
	var t = epoch.Time()
	isValid = t.After(minTime) && t.Before(maxTime)

	return
}

const (
	// the number of nanosecond per second 1e9 int64
	nanoPerSecond = int64(time.Second / time.Nanosecond)
	// epoch value representing zero time
	epochZeroTime = Epoch(math.MinInt64)
	// epoch value representing invalid
	epochInvalid = Epoch(math.MinInt64 + 1)
)

var (
	// minTime is the minimum allowed time
	minTime = time.Date(1677, 12, 29, 0, 0, 0, 0, time.UTC)
	// maxTime is the maximum allowed time
	maxTime = time.Date(2262, 1, 3, 0, 0, 0, 0, time.UTC)
)

// func (t time.Time) UnixNano() int64
//   - the number of nanoseconds elapsed since January 1, 1970 UTC
//   - The result is undefined if the Unix time in nanoseconds cannot be
//     represented by an int64 (a date before the year 1678 or after 2262)
//   - UnixNano on the zero Time is undefined
//   - The result does not depend on the location associated with t
var _ = (&time.Time{}).UnixNano()

// func time.Unix(sec int64, nsec int64) time.Time
//   - Unix returns the local Time corresponding to the given Unix time,
//     sec seconds and nsec nanoseconds since January 1, 1970 UTC.
//     It is valid to pass nsec outside the range [0, 999999999].
//     Not all sec values have a corresponding time value.
//     One such value is 1<<63-1 (the largest int64 value).
var _ = time.Unix

// var time.Local *time.Location
//   - Local represents the system's local time zone
var _ = time.Local

// var time.UTC *time.Location
//   - UTC represents Universal Coordinated Time (UTC)
var _ = time.UTC
