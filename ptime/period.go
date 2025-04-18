/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

const (
	// fraction subdivides a period
	// maximum bits are used: 2^63 is 100% of period
	FractionScale = uint64(1) << 63
)

// tPeriodEpoch is epoch for all time periods
var tPeriodEpoch = time.Now()

type PeriodIndex uint64

// Period provides real-time based fixed-length interval perioding ordered using a uint64 zero-based index.
//   - Period provides first period index and fractional usage of the first period
type Period struct {
	// interval is ns duration of a period
	interval time.Duration
	// period0 is the numeric value for the first period
	//	- the exact number is relative to a process-wide timestamp
	period0 PeriodIndex
	// this Period was instantiated sometime during the period period0
	//	- fraction0 is how much the first period0 is prior to instantiation
	//	- fraction0 0 means the none of period0 was active
	//	- fraction0 2^63 means all of period0 was active
	fraction0 uint64
}

// NewPeriod returns a new numbered-interval sequence.
//   - interval: 1ns to 24 hours
func NewPeriod(interval time.Duration, fieldp ...*Period) (period *Period) {
	if interval == 0 {
		panic(perrors.NewPF("interval cannot be zero"))
	} else if interval > 24*time.Hour {
		panic(perrors.NewPF("interval cannot be longer than 24 hours"))
	}

	if len(fieldp) > 0 {
		period = fieldp[0]
	}
	if period == nil {
		period = &Period{}
	}

	var t = time.Now()
	*period = Period{
		interval: interval,
	}
	period.period0 = period.Index(t)

	// calculate what fraction of the first period is active
	// uint64 valid decimal digits is : 64 * log10(2) ≈ 19
	// use scale factor that is power of 2: 2^63
	var t0 = t.Truncate(interval)
	var inactiveDuration = t.Sub(t0)
	period.fraction0 = FractionScale - uint64(float64(inactiveDuration)/float64(interval)*float64(FractionScale))

	return
}

// Index returns the index number for the current period or the period at time t
//   - t: optional timestamp, default now
//   - — timestamp cannot be prior to [NewPeriod] invocation
//   - index: number p.period0 or larger
func (p *Period) Index(t ...time.Time) (index PeriodIndex) {
	if p.interval == 0 {
		panic(perrors.NewPF("uninitialized period"))
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
//   - now: a later index, cannot be less than period0 or before
//   - before: an earlier index, cannot be less than period0
//   - periods: >= 0
//   - —
//   - now and before are obtained from [Period.Index] or [Period.Sub]
func (p *Period) Since(now, before PeriodIndex) (periods int) {
	if p.interval == 0 {
		panic(perrors.NewPF("uninitialized period"))
	} else if before < p.period0 {
		panic(perrors.ErrorfPF("Period.Sub with before less than period0: %d now: %d period0: %d",
			before, now, p.period0))
	} else if now < before {
		panic(perrors.ErrorfPF("Period.Sub with before greater than now: %d now: %d period0: %d",
			before, now, p.period0))
	}

	periods = int(now - before)

	return
}

// Sub returns a previous period index by n intervals
//   - now: cannot be less than period0
//   - n: cannot be negative
//   - periodIndex: will not be less than period0
//   - —
//   - now is obtained from [Period.Index] or [Period.Sub]
func (p *Period) Sub(now PeriodIndex, n int) (periodIndex PeriodIndex) {
	if p.interval == 0 {
		panic(perrors.NewPF("uninitialized period"))
	} else if now < p.period0 {
		panic(perrors.ErrorfPF("Period.Sub with now less than period0: %d %d", now, p.period0))
	} else if n < 0 {
		panic(perrors.ErrorfPF("Period.Sub with n negative: %d", n))
	}

	// zero periods
	if n == 0 {
		periodIndex = now
		return // nothing to do return
	}

	// maxN is the n value that returns period0
	var maxN = int(now - p.period0)
	if n > maxN {
		n = maxN
	}
	periodIndex = now - PeriodIndex(n)

	return
}

// Available returns the correct number of slice entries to include now
//   - now: cannot be less than period0
//   - periods: length of range from period0 to now, but no larger than cap
//   - — value: 1…cap
//   - — after the inital few periods, always returns cap
//   - —
//   - now is obtained from [Period.Index] or [Period.Sub]
func (p *Period) Available(now PeriodIndex, cap int) (periods int) {
	if p.interval == 0 {
		panic(perrors.NewPF("uninitialized period"))
	} else if now < p.period0 {
		panic(perrors.ErrorfPF("Period.Sub with now less than period0: %d %d", now, p.period0))
	}
	if periods = int(now-p.period0) + 1; periods > cap {
		periods = cap
	}
	return
}
