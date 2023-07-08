/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Period provides real-time based fixed-length interval perioding ordered using a uint64 zero-based index.
package ptime

import (
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

const (
	FractionScale = uint64(1) << 63
)

// tPeriodEpoch is epoch for all time periods
var tPeriodEpoch = time.Now()

type PeriodIndex uint64

// Period provides real-time based fixed-length interval perioding ordered using a uint64 zero-based index.
//   - Period provides first period index and fractional usage of the first period
type Period struct {
	interval  time.Duration
	period0   PeriodIndex
	fraction0 uint64
}

// NewPeriod returns a new numbered-interval sequence.
func NewPeriod(interval time.Duration) (period *Period) {
	t := time.Now()
	p := Period{interval: interval}
	p.period0 = p.Index(t)

	// calculate what fraction of the first period is active
	// uint64 valid decimal digits is : 64 * log10(2) ≈ 19
	// use scale factor that is power of 2: 2^63
	t0 := t.Truncate(interval)
	inactiveDuration := t.Sub(t0)
	p.fraction0 = FractionScale - uint64(float64(inactiveDuration)/float64(interval)*float64(FractionScale))

	return &p
}

// Index returns the index number for the current period or the period at time t
//   - Index is zero-based
func (p *Period) Index(t ...time.Time) (index PeriodIndex) {
	if p.interval <= 0 {
		panic(perrors.ErrorfPF("period must be positive: %s", Duration(p.interval)))
	}

	// get the time for which to get index
	var t0 time.Time
	if len(t) > 0 {
		t0 = t[0]
	}
	if t0.IsZero() {
		t0 = time.Now()
	}
	if t0.Before(tPeriodEpoch) {
		panic(perrors.ErrorfPF("time before epoch: %s %s", t0.Format(cyclebreaker.Rfc3339ns), tPeriodEpoch.Format(cyclebreaker.Rfc3339ns)))
	}

	// get index
	if index = PeriodIndex(t0.Sub(tPeriodEpoch) / p.interval); index < p.period0 {
		panic(perrors.ErrorfPF("time before period0: index: %d period0: %d %s epoch: %s",
			index, p.period0,
			t0.Format(cyclebreaker.Rfc3339ns), tPeriodEpoch.Format(cyclebreaker.Rfc3339ns)),
		)
	}
	return
}

// Since returns the number of periods difference that now is greater than before
//   - periods is >= 0
//   - now and before must be at least p.period0
//   - before cannot be greater than now
func (p *Period) Since(now, before PeriodIndex) (periods int) {
	if before < p.period0 {
		panic(perrors.ErrorfPF("Period.Sub with before less than period0: %d now: %d period0: %d",
			before, now, p.period0))
	} else if now < before {
		panic(perrors.ErrorfPF("Period.Sub with before greater than now: %d now: %d period0: %d",
			before, now, p.period0))
	}

	periods = int(now - before)

	return
}

// Sub returns a past period index by n intervals
//   - periodIndex will not be less than p.period0
//   - now cannot be less than p.period0
func (p *Period) Sub(now PeriodIndex, n int) (periodIndex PeriodIndex) {
	if now < p.period0 {
		panic(perrors.ErrorfPF("Period.Sub with now less than period0: %d %d", now, p.period0))
	} else if n == 0 {
		return // nothing to do return
	} else if n < 0 {
		panic(perrors.ErrorfPF("Period.Sub with n negative: %d", n))
	} else if maxN := int(now - p.period0); n > maxN {
		n = maxN
	}

	periodIndex = now - PeriodIndex(n)

	return
}

// Available returns the number of possible periods 1… but no greater than cap
//   - now cannot be less than period0
func (p *Period) Available(now PeriodIndex, cap int) (periods int) {
	if now < p.period0 {
		panic(perrors.ErrorfPF("Period.Sub with now less than period0: %d %d", now, p.period0))
	}
	if periods = int(now-p.period0) + 1; periods > cap {
		periods = cap
	}
	return
}
