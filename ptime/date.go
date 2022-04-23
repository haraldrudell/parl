/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import "time"

// ptime.Date is like time.Date with a flag valid that indicates
// whether the provided values represents a valid date.
// year is a 4-digit year, ie. 2022 is 2022.
// month 1–12, day 1–31, hour 0–23 minute 0-59, sec 0–59
func Date(year int, month int, day int, hour int, min int, sec int, nsec int, loc *time.Location) (t time.Time, valid bool) {
	t = time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc)
	valid = t.Year() == year && t.Month() == time.Month(month) && t.Day() == day &&
		t.Hour() == hour && t.Minute() == min && t.Second() == sec &&
		t.Nanosecond() == nsec
	return
}

// ptime.Date is like time.Date with a flag valid that indicates
// whether the provided values represents a valid date.
// year is a 4-digit year, ie. 2022 is 2022.
// month 1–12, day 1–31, hour 0–23 minute 0-59, sec 0–59
// A year below 1000 is not considered valid
func Date1k(year int, month int, day int, hour int, min int, sec int, nsec int, loc *time.Location) (t time.Time, valid bool) {
	t, valid = Date(year, month, day, hour, min, sec, nsec, loc)
	valid = valid && year >= 1000
	return
}
