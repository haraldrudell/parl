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

// HaltReport is a value object representing a detected Go runtime execution halt
type HaltReport struct {
	// report number 1…
	N int
	// when halt started
	T time.Time
	// halt duration
	D time.Duration
}

func (r *HaltReport) String() (s string) {
	return parl.Sprintf("%s %d %s",
		parl.ShortSpace(r.T),
		r.N,
		ptime.Duration(r.D),
	)
}
