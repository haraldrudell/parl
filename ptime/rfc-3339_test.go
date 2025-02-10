/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package parltime provides on-time timers, 64-bit epoch, formaatting and other time functions.
package ptime

import (
	"reflect"
	"testing"
	"time"
)

func TestRfc3339(t *testing.T) {
	//t.Error("Logging on")
	var (
		t0      = time.Date(2025, 12, 31, 1, 2, 3, 123456789, time.UTC)
		t0Local = t0.Local().Format(rfc3339)
	)
	const (
		t0UTC = "2025-12-31 01:02:03+00:00"
	)
	t.Logf("t0UTC: %s", t0UTC)
	t.Logf("t0Local: %s", t0Local)
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"zero-time", args{time.Time{}}, "0001-01-01 00:00:00+00:00"},
		{"UTC time", args{t0}, t0UTC},
		{"local time", args{t0.Local()}, t0Local},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Rfc3339(tt.args.t); got != tt.want {
				t.Errorf("Rfc3339() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMs(t *testing.T) {
	//t.Error("Logging on")
	const (
		d  = time.Duration(int64(0.123456789 * float32(time.Second)))
		d2 = time.Second + 200*time.Millisecond
	)
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"0.123456789s", args{d}, "123ms"},
		{"1.2s", args{d2}, "1.2s"},
		{"zero", args{0}, "0s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Ms(tt.args.d); got != tt.want {
				t.Errorf("Ms() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	//t.Error("Logging on")
	const (
		errorYes = true
		errorNo  = false
	)
	var (
		// 30 minutes east of UTC
		customLocation = time.FixedZone("", 30*60)
		// rfc 3339 string-date in UTC time zone matching time.Time tUTCExp
		tUTC = "2025-12-31 01:02:03+00:00"
		// time.Time matching string tUTC
		tUTCExp = time.Date(2025, 12, 31, 1, 2, 3, 0, time.FixedZone("", 0))
		// A UTC time string using Z that fails rfc 3339 parsing
		tUTCZ = "2025-12-31 01:02:03Z"
		// a time string without time offset that fails rfc 3339 parsing
		tNoTz = "2025-12-31 01:02:03"
		// an rfc 3339 time string in time offset customLocation matching time.Time tCustomExp
		tCustomTz = "2025-12-31 01:02:03+00:30"
		// time.Time matching string tCustomTz
		tCustomExp = time.Date(2025, 12, 31, 1, 2, 3, 0, customLocation)
		// rfc 3339 time string in local time zone
		//	- “2025-12-30 17:02:03-08:00”
		tLocal = time.Date(2025, 12, 31, 1, 2, 3, 123456789, time.UTC).
			Local().Format(rfc3339)
		tLocalExp = tUTCExp.Local()
	)
	t.Logf("tLocal: %s", tLocal)
	t.Logf("customLocation: %+v", *customLocation)
	var tzx, _ = ParseTime(tLocal)
	t.Logf("ParseTime tLocal time zone is Local: %t", tzx.Location() == time.Local)
	tzx, _ = ParseTime(tUTC)
	t.Logf("ParseTime tUTC time zone: “%s” %v", tzx.Location(), *tzx.Location())
	type args struct {
		dateString string
	}
	tests := []struct {
		name    string
		args    args
		wantTm  time.Time
		wantErr bool
	}{
		{"UTC", args{tUTC}, tUTCExp, errorNo},
		{"UTC-Z", args{tUTCZ}, time.Time{}, errorYes},
		{"no time-zone", args{tNoTz}, time.Time{}, errorYes},
		{"custom time-zone", args{tCustomTz}, tCustomExp, errorNo},
		{"local time-zone", args{tLocal}, tLocalExp, errorNo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTm, err := ParseTime(tt.args.dateString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTm, tt.wantTm) {
				t.Errorf("ParseTime() = “%v”, want “%v”", gotTm, tt.wantTm)
			}
		})
	}
}
