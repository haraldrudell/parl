/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package halt

import (
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/ptime"
)

// HaltReport is a value type representing a detected Go runtime execution halt
type HaltReport struct {
	// report number 1…
	Number int
	// when halt started
	Timestamp time.Time
	// halt duration
	Duration time.Duration
}

func (r *HaltReport) String() (s string) {
	return parl.Sprintf("%s %d %s",
		parl.ShortSpace(r.Timestamp),
		r.Number,
		ptime.Duration(r.Duration),
	)
}
