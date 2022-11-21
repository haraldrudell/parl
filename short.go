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

// Short provides exact world time in compact second-precision: 060102_15:04:05Z07
// the time printed is either tim or time.Now().
func Short(tim ...time.Time) (s string) {
	return timeAndFormat(tim, shortHour, shortMinute)
}

// Short provides exact world time in compact second-precision: 060102 15:04:05Z07
// the time printed is either tim or time.Now().
func ShortSpace(tim ...time.Time) (s string) {
	return timeAndFormat(tim, shortHourSpace, shortMinuteSpace)
}

// ShortMs provides exact world time in compact milli-second-precision: 060102_15:04:05Z07
// the time printed is either tim or time.Now().
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
