/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"

	"github.com/haraldrudell/parl/perrors"
)

const (
	// RFC 3339 (email) time format
	rfc3339 = "2006-01-02 15:04:05-07:00"
	// RFC3339NanoSpace RFC3339 format, ns precision, space separator
	RFC3339NanoSpace string = "2006-01-02 15:04:05.999999999Z07:00"
)

// Rfc3339 converts time to string with second precision and time offset
//   - “2006-01-02 15:04:05-07:00”
//   - “2025-12-31 01:02:03+00:00”
//   - t: second-precision
//   - local time zone or UTC by t.Location, it’s built-in time zone information
//   - typically, time.Time is in t.Local
//   - [S] is same in UTC
func Rfc3339(t time.Time) string { return t.Format(rfc3339) }

// ParseTime parses output from [Rfc3339], second precision and time offset
//   - layout: “2006-01-02 15:04:05-07:00”
//   - tm Location depends on dateString time offset
//   - — time offset matching time.Local uses this location
//   - — other time offset returns custom time zone
//   - ‘Z’ for time zone or missing time zone is error
func ParseTime(dateString string) (tm time.Time, err error) {
	if tm, err = time.Parse(rfc3339, dateString); err != nil {
		err = perrors.Errorf("time.Parse: '%w'", err)
	}
	return
}

// Ns converts time to string with nanosecond accuracy and UTC time zone
//   - “2025-12-31 01:02:03.123456789+00:00”
//   - parsed by [ParseNs]
func Ns(t time.Time) string {
	return t.UTC().Truncate(time.Nanosecond).Format(time.RFC3339Nano)
}

// ParseNs parses an RFC3339 time string with nanosecond accuracy
//   - “2025-12-31 01:02:03.123456789+00:00”
//   - output by [Ns]
func ParseNs(timeString string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC3339Nano, timeString)
	if err != nil {
		t, err = time.Parse(RFC3339NanoSpace, timeString)
	}
	return
}

// S converts time to string with second accuracy and UTC location
//   - “2025-12-31 01:02:03.123456789+00:00”
//   - similar to [Rfc3339] [NsLocal]
func S(t time.Time) string {
	return t.UTC().Truncate(time.Nanosecond).Format(time.RFC3339)
}

// NsLocal converts time to string with nanosecond accuracy and local time zone
//   - “2025-12-31 01:02:03.123456789-08:00”
//   - similar to [Rfc3339] [S]
func NsLocal(t time.Time) string {
	return t.Local().Truncate(time.Nanosecond).Format(time.RFC3339Nano)
}

// GetTimeString rfc 3339: email time format
//   - default: time.Now() with local time offset
//   - s: “2020-12-04 20:20:17-08:00”
func GetTimeString(wallTime time.Time) (s string) {
	var when time.Time
	if !wallTime.IsZero() {
		when = wallTime
	} else {
		when = time.Now()
	}
	s = when.Format(rfc3339)
	return
}
