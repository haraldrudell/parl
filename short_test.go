/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
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

func TestShort(t *testing.T) {
	exp := "220101_00:00:00-08"
	input := timeFixture
	actual := Short(input)

	if actual != exp {
		t.Errorf("ptime.Short %q exp %q", actual, exp)
	}
}
