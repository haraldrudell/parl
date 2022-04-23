/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
	"time"
)

const (
	IANASweetHomeSanFrancisco = "America/Los_Angeles"
	timeZonePST               = "PST"
	offsetPSTh                = -8
	exp                       = ""
)

func TestRfc3339nsz(t *testing.T) {
	var locCalifornia *time.Location
	var err error

	// get a known time for a klnown location
	if locCalifornia, err = time.LoadLocation(IANASweetHomeSanFrancisco); err != nil {
		t.Errorf("time.LoadLocation err: %v", err)
	}
	tim := time.Date(2022, time.Month(1), 1, 0, 0, 0, 0, locCalifornia)

	// verify time zone
	var name string
	var offsetS int
	name, offsetS = tim.Zone()
	if name != timeZonePST {
		t.Errorf("time zone abbreviation: %s exp %s", name, timeZonePST)
	}
	expOffset := offsetPSTh * int(time.Hour/time.Second)
	if offsetS != expOffset {
		t.Errorf("time zone abbreviation: %d exp %d", offsetS, expOffset)
	}

	//actual := tim.Format(Rfc3339nsz)
	//if actual != exp {
	// It’s in local time zone, stupid
	// this does not actually work… t.Errorf(".Format(parl.Rfc3339nsz) format: %q actual: %s exp: %s", Rfc3339nsz, actual, exp)
	// use ptime.Rfc3339nsz
	//}
}
