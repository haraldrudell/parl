/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import (
	"reflect"
	"testing"
	"time"
)

func TestDATETIMEtoTime(t *testing.T) {
	const noErr = false

	type args struct {
		sqliteText string
	}
	tests := []struct {
		name    string
		args    args
		wantT   time.Time
		wantErr bool
	}{
		{"timestamp", args{"2019-10-19 22:26:34.000 +00:00"}, time.Date(2019, 10, 19, 22, 26, 34, 0, time.UTC).Local(), noErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotT, err := DATETIMEtoTime(tt.args.sqliteText)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLiteDATETIMEtoTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotT, tt.wantT) {
				t.Errorf("SQLiteDATETIMEtoTime() = %v, want %v", gotT, tt.wantT)
			}
		})
	}
}

func TestTimeToDATETIME(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name         string
		args         args
		wantDateTime string
	}{
		{"time", args{time.Date(2019, 10, 19, 22, 26, 34, 0, time.UTC).Local()}, "2019-10-19 22:26:34.000 +00:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDateTime := TimeToDATETIME(tt.args.t); gotDateTime != tt.wantDateTime {
				t.Errorf("TimeToDATETIME() = %v, want %v", gotDateTime, tt.wantDateTime)
			}
		})
	}
}
