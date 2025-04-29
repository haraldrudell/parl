/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"time"

	"github.com/haraldrudell/parl"
)

type Rate struct {
	// String()
	RateType
	delta    uint64
	duration time.Duration
	hasValue bool
}

// Rate is [parl.Rate]
var _ parl.Rate = &Rate{}

func (rt *Rate) Clone() (rate parl.Rate) {
	var r2 Rate = *rt
	return &r2
}
func (rt *Rate) Delta() (delta uint64)              { return rt.delta }
func (rt *Rate) Duration() (duration time.Duration) { return rt.duration }
func (rt *Rate) HasValue() (hasValue bool)          { return rt.hasValue }
