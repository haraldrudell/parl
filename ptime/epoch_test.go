/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestEpoch_Time(t *testing.T) {
	//
	// Time converts Epoch to time.Time in time.Local
	//
	const (
		// the number of nanosecond per second 1e9 int64
		nsS = int64(time.Second / time.Nanosecond)
	)
	var (
		epochMin = Epoch(math.MinInt64)
		minTime  = time.Unix(int64(epochMin)/nsS, int64(epochMin)%nsS)

		epochMax = Epoch(math.MaxInt64)
		maxTime  = time.Unix(int64(epochMax)/nsS, int64(epochMax)%nsS)

		epochZero = Epoch(0)
		janUTC    = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).Local()

		minusOneEpoch = Epoch(-time.Second - time.Nanosecond)
		mnusOneTime   = time.Date(1969, 12, 31, 23, 59, 58, 999999999, time.UTC).Local()
	)

	// minTime 1677-09-20 16:19:45.145224192 -0752 LMT
	t.Logf("minTime %s", minTime.String())
	// maxTime 2262-04-11 16:47:16.854775807 -0700 PDT
	t.Logf("maxTime %s", maxTime.String())

	_ = janUTC
	tests := []struct {
		name  string
		epoch Epoch
		wantT time.Time
	}{
		{"January 1, 1970 UTC", epochZero, janUTC},
		{"zero-time", epochZeroTime, time.Time{}},
		{"invalid", epochInvalid, time.Time{}},
		{"minusOne", minusOneEpoch, mnusOneTime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotT := tt.epoch.Time(); !reflect.DeepEqual(gotT, tt.wantT) {
				t.Errorf("Epoch.Time() = %v, want %v", gotT, tt.wantT)
			}
		})
	}
}

func TestEpochNow(t *testing.T) {
	//
	// EpochNow converts time.Time to Epoch
	//
	var (
		zeroTime = []time.Time{{}}

		janUTC      = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		epochJanUTC = Epoch(0)

		janUTCOneSecond = time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC)
		epochOneSecond  = Epoch(time.Second)

		year1000 = time.Date(1000, 1, 1, 0, 0, 1, 0, time.UTC)
		year3000 = time.Date(1000, 1, 1, 0, 0, 1, 0, time.UTC)
	)
	type args struct {
		t []time.Time
	}
	tests := []struct {
		name      string
		args      args
		wantEpoch Epoch
	}{
		{"zero-value", args{zeroTime}, epochZeroTime},
		{"January 1, 1970 UTC", args{[]time.Time{janUTC}}, epochJanUTC},
		{"Epoch one-second", args{[]time.Time{janUTCOneSecond}}, epochOneSecond},
		{"year 1000", args{[]time.Time{year1000}}, epochInvalid},
		{"year 3000", args{[]time.Time{year3000}}, epochInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotEpoch := EpochNow(tt.args.t...); gotEpoch != tt.wantEpoch {
				t.Errorf("EpochNow() = %v, want %v", gotEpoch, tt.wantEpoch)
			}
		})
	}

	// EpochNow without argument should use time.Now
	//   - [time.Now] on macOS is μs precision
	var before = Epoch(time.Now().UnixNano())
	<-time.NewTimer(2 * time.Microsecond).C
	var actual = EpochNow()
	<-time.NewTimer(2 * time.Microsecond).C
	var after = Epoch(time.Now().UnixNano())
	if actual <= before {
		t.Errorf("EpochNow no argument: too early %d <= %d", actual, before)
	}
	if actual >= after {
		t.Errorf("EpochNow no argument: too late %d >= %d", actual, after)
	}
}
