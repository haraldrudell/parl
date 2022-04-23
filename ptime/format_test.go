/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

const (
	IANASweetHomeSanFrancisco = "America/Los_Angeles"
	timeZonePST               = "PST"
	offsetPSTh                = -8
	exp                       = "2022-01-01T08:00:00.000000000Z"
	expuS                     = "2022-01-01T08:00:00.000000Z"
	expmS                     = "2022-01-01T08:00:00.000Z"
	expS                      = "2022-01-01T08:00:00Z"
)

var locCaliforniaFixture = func() (loc *time.Location) {
	var err error

	// get a known time for a known location
	if loc, err = time.LoadLocation(IANASweetHomeSanFrancisco); err != nil {
		panic(perrors.Errorf("time.LoadLocation: %w", err))
	}
	return loc
}()

var timeFixture = func() (t time.Time) {
	t = time.Date(2022, time.Month(1), 1, 0, 0, 0, 0, locCaliforniaFixture)

	// verify time zone
	var name string
	var offsetS int
	name, offsetS = t.Zone()
	if name != timeZonePST {
		panic(perrors.Errorf("time zone abbreviation: %s exp %s", name, timeZonePST))
	}
	expOffset := offsetPSTh * int(time.Hour/time.Second)
	if offsetS != expOffset {
		panic(perrors.Errorf("time zone abbreviation: %d exp %d", offsetS, expOffset))
	}
	return t
}()

func TestRfc3339nsz(t *testing.T) {
	actual := Rfc3339nsz(timeFixture)
	if actual != exp {
		t.Errorf(".Format(parl.Rfc3339nsz) format: %q actual: %s exp: %s", rfc3339nsz, actual, exp)
	}
}

func TestRfc3339usz(t *testing.T) {
	actual := Rfc3339usz(timeFixture)
	if actual != expuS {
		t.Errorf(".Format(parl.Rfc3339usz) format: %q actual: %s exp: %s", rfc3339usz, actual, expuS)
	}
}

func TestRfc3339msz(t *testing.T) {
	actual := Rfc3339msz(timeFixture)
	if actual != expmS {
		t.Errorf(".Format(parl.Rfc3339msz) format: %q actual: %s exp: %s", rfc3339msz, actual, expmS)
	}
}

func TestRfc3339sz(t *testing.T) {
	actual := Rfc3339sz(timeFixture)
	if actual != expS {
		t.Errorf(".Format(parl.Rfc3339sz) format: %q actual: %s exp: %s", rfc3339sz, actual, expS)
	}
}

func TestShort(t *testing.T) {
	exp := "220101_00:00:00-08"
	input := timeFixture
	actual := Short(input)

	if actual != exp {
		t.Errorf("ptime.Short %q exp %q", actual, exp)
	}
}
