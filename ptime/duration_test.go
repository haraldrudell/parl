/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"math"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name           string
		args           args
		wantPrintableD string
	}{
		{"ns", args{time.Duration(math.Round(math.Pi * 100))}, "314ns"},            // 314ns
		{"µs", args{time.Duration(math.Round(math.Pi * 1e5))}, "314µs"},            // 314.159µs
		{"ms", args{time.Duration(math.Round(math.Pi * 1e8))}, "314ms"},            // 314.159265ms
		{"10s", args{time.Duration(math.Round(math.Pi * 1e9))}, "3.1s"},            // 3.141592654s
		{"s", args{time.Duration(math.Round(math.Pi * 1e10))}, "31s"},              // 31.415926536s
		{"min", args{time.Duration(math.Round(math.Pi * 1e11))}, "5m14s"},          // 5m14.159265359s
		{"h", args{time.Duration(math.Round(math.Pi * 2e13))}, "17h27m"},           // 17h27m11.853071796s
		{"days", args{time.Duration(math.Round(math.Pi * 1e14))}, "3d15h"},         // 87h15m59.265358979s
		{"months", args{time.Duration(math.Round(math.Pi * 1e15))}, "36d"},         // 872h39m52.653589793s
		{"years", args{time.Duration(math.Round(math.Pi * 1e17))}, "3636d"},        // 87266h27m45.358979328s
		{"negative", args{time.Duration(-time.Second / 3)}, "-333ms"},              // 87266h27m45.358979328s
		{"negative mins", args{-time.Hour - time.Minute - time.Second}, "-1h1m1s"}, // 87266h27m45.358979328s
		{"negative hours", args{-25 * time.Hour}, "-1d1h"},                         // 87266h27m45.358979328s
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Duration(tt.args.d)
			t.Logf("%s Duration(%s) → %s", tt.name, tt.args.d.String(), actual)
			if actual != tt.wantPrintableD {
				t.Errorf("Duration() = %q, want %q", actual, tt.wantPrintableD)
				t.Fail()
			}
		})
	}
}

func TestDurationHMS(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name             string
		args             args
		wantPrintableHMS string
	}{
		{"zero", args{0}, "00:00:00"},
		{"-ns", args{-time.Nanosecond}, "-00:00:00"},
		{"-s", args{-time.Second}, "-00:00:01"},
		{"240h", args{240 * time.Hour}, "240:00:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPrintableHMS := DurationHMS(tt.args.d); gotPrintableHMS != tt.wantPrintableHMS {
				t.Errorf("DurationHMS() = %v, want %v", gotPrintableHMS, tt.wantPrintableHMS)
			}
		})
	}
}
