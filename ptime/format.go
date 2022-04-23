/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"
)

const (
	rfc3339sz         = "2006-01-02T15:04:05Z"
	rfc3339msz        = "2006-01-02T15:04:05.000Z"
	rfc3339usz        = "2006-01-02T15:04:05.000000Z"
	rfc3339nsz        = "2006-01-02T15:04:05.000000000Z"
	shortHour         = "060102_150405Z07"
	shortMinute       = "060102_150405Z0700"
	offsetHourDivisor = int(time.Hour / time.Second)
)

// Rfc3339nsz prints a time using UTC and nanoseconds precision.
//  "2022-01-01T08:00:00.000000000Z"
func Rfc3339nsz(t time.Time) (s string) {
	// this must be a function because of the .UTC incovation
	return t.UTC().Format(rfc3339nsz)
}

// Rfc3339usz prints a time using UTC and microseconds precision.
//  "2022-01-01T08:00:00.000000Z"
func Rfc3339usz(t time.Time) (s string) {
	// this must be a function because of the .UTC incovation
	return t.UTC().Format(rfc3339usz)
}

// Rfc3339msz prints a time using UTC and milliseconds precision.
//  "2022-01-01T08:00:00.000Z"
func Rfc3339msz(t time.Time) (s string) {
	// this must be a function because of the .UTC incovation
	return t.UTC().Format(rfc3339msz)
}

// Rfc3339sz prints a time using UTC and seconds precision.
//  "2022-01-01T08:00:00Z"
func Rfc3339sz(t time.Time) (s string) {
	// this must be a function because of the .UTC incovation
	return t.UTC().Format(rfc3339sz)
}

// ParseRfc3339nsz parses a UTC time string at nanoseconds precision.
//  "2022-01-01T08:00:00.000000000Z"
func ParseRfc3339nsz(timeString string) (t time.Time, err error) {
	t, err = time.Parse(rfc3339nsz, timeString)
	return
}

func Short(tim ...time.Time) (s string) {

	// ensure t holds a time
	var t time.Time
	if len(tim) > 0 {
		t = tim[0]
	}
	if t.IsZero() {
		t = time.Now()
	}

	// pick layout using Zone.offset
	layout := shortHour
	if _, offsetS := t.Zone(); offsetS%offsetHourDivisor != 0 {
		layout = shortMinute
	}

	return t.Format(layout)
}
