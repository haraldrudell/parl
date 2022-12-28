/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "time"

const (
	shortHour         = "060102_15:04:05Z07"
	shortMinute       = "060102_15:04:05Z0700"
	shortHourSpace    = "060102 15:04:05Z07"
	shortMinuteSpace  = "060102 15:04:05Z0700"
	offsetHourDivisor = int(time.Hour / time.Second)
	msHour            = "060102_15:04:05.000Z07"
	msMinute          = "060102_15:04:05.000Z0700"
)

// Short provides a brief time-stamp in compact second-precision including time-zone.
//   - sample: 060102_15:04:05-08
//   - The timestamp does not contain space.
//   - time zone is what is included in tim, typically time.Local
//   - if tim is not specified, time.Now() in local time zone
func Short(tim ...time.Time) (s string) {
	return timeAndFormat(tim, shortHour, shortMinute)
}

// Short provides a brief time-stamp in compact second-precision including time-zone.
//   - sample: 060102 15:04:05-08
//   - date is 6-digit separated from time by a space
//   - time zone is what is included in tim, typically time.Local
//   - if tim is not specified, time.Now() in local time zone
func ShortSpace(tim ...time.Time) (s string) {
	return timeAndFormat(tim, shortHourSpace, shortMinuteSpace)
}

// ShortMs provides a brief time-stamp in compact milli-second-precision including time-zone.
//   - sample: 060102_15:04:05.123-08
//   - The timestamp does not contain space.
//   - time zone is what is included in tim, typically time.Local
//   - if tim is not specified, time.Now() in local time zone
func ShortMs(tim ...time.Time) (s string) {
	return timeAndFormat(tim, msHour, msMinute)
}

func timeAndFormat(tim []time.Time, hour, minute string) (s string) {

	// ensure t holds a time
	var t time.Time
	if len(tim) > 0 {
		t = tim[0]
	}
	if t.IsZero() {
		t = time.Now()
	}

	// pick layout using Zone.offset
	var format string
	if _, offsetS := t.Zone(); offsetS%offsetHourDivisor != 0 {
		format = minute
	} else {
		format = hour
	}

	return t.Format(format)
}
