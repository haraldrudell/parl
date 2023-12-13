/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package parltime provides on-time timers, 64-bit epoch, formaatting and other time functions.
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

// Rfc3339 converts local time to string with second precision and time offset: 2006-01-02 15:04:05-07:00
func Rfc3339(t time.Time) string {
	return t.Format(rfc3339)
}

// ParseTime parses output from Rfc3339
func ParseTime(dateString string) (tm time.Time, err error) {
	if tm, err = time.Parse(rfc3339, dateString); err != nil {
		err = perrors.Errorf("time.Parse: '%w'", err)
	}
	return
}

// Ms gets duration in milliseconds
func Ms(d time.Duration) string {
	return d.Truncate(time.Millisecond).String()
}

// Ns converts time to string with nanosecond accuracy and UTC location
func Ns(t time.Time) string {
	return t.UTC().Truncate(time.Nanosecond).Format(time.RFC3339Nano)
}

// ParseNs parses an RFC3339 time string with nanosecond accuracy
func ParseNs(timeString string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC3339Nano, timeString)
	if err != nil {
		t, err = time.Parse(RFC3339NanoSpace, timeString)
	}
	return
}

// S converts time to string with second accuracy and UTC location
func S(t time.Time) string {
	return t.UTC().Truncate(time.Nanosecond).Format(time.RFC3339)
}

// ParseS parses an RFC3339 time string with nanosecond accuracy
func ParseS(timeString string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC3339, timeString)
	return
}

// NsLocal converts time to string with nanosecond accuracy and local time zone
func NsLocal(t time.Time) string {
	return t.Local().Truncate(time.Nanosecond).Format(time.RFC3339Nano)
}

// SGreater compares two times first rouding to second
func SGreater(t1 time.Time, t2 time.Time) bool {
	return t1.Truncate(time.Second).After(t2.Truncate(time.Second))
}

// GetTimeString rfc 3339: email time format 2020-12-04 20:20:17-08:00
func GetTimeString(wallTime *time.Time) (s string) {
	var when time.Time
	if wallTime != nil {
		when = *wallTime
	} else {
		when = time.Now()
	}
	s = when.Format(rfc3339)
	return
}
