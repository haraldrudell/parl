/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

// tPeriodEpoch is epoch for all time periods
var tPeriodEpoch = time.Now()

type PeriodIndex uint64

type Period struct {
	interval time.Duration
}

func NewPeriod(interval time.Duration) (period *Period) {
	if interval <= 0 {
		panic(perrors.ErrorfPF("period must be positive: %s", ptime.Duration(interval)))
	}
	return &Period{interval: interval}
}

// Index returns the index number for the current period or the period at time t
//   - Index is zero-based
func (p *Period) Index(t ...time.Time) (index PeriodIndex) {

	// get the time for which to get index
	var t0 time.Time
	if len(t) > 0 {
		t0 = t[0]
	}
	if t0.IsZero() {
		t0 = time.Now()
	}
	if t0.Before(tPeriodEpoch) {
		panic(perrors.ErrorfPF("time before epoch: %s %s", t0.Format(parl.Rfc3339ns), tPeriodEpoch.Format(parl.Rfc3339ns)))
	}

	// get index
	index = PeriodIndex(t0.Sub(tPeriodEpoch) / p.interval)
	return
}
