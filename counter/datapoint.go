/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
)

// Datapoint tracks a fluctuating value with average. Thread-safe.
type Datapoint struct {
	period time.Duration

	lock  sync.Mutex
	value uint64
	max   uint64
	min   uint64
	// period in in progress
	periodStart  time.Time
	isFullPeriod bool
	nPeriod      uint64
	aggregate    float64
	// valid values
	average float64
	n       uint64
}

var _ parl.DatapointValue = &Datapoint{}

func newDatapoint(period time.Duration) (datapoint parl.Datapoint) {
	if period <= 0 {
		panic(perrors.ErrorfPF("period must be positive: %s", ptime.Duration(period)))
	}
	return &Datapoint{period: period}
}

// SetValue records a new datapoint value
func (dt *Datapoint) SetValue(value uint64) (datapoint parl.Datapoint) {
	tNow := time.Now() // best possible tNow
	datapoint = dt
	dt.lock.Lock()
	defer dt.lock.Unlock()

	// update value max min
	dt.value = value
	isFirst := dt.periodStart.IsZero()
	if isFirst || value > dt.max {
		dt.max = value
	}
	if isFirst || value < dt.min {
		dt.min = value
	}

	// determine period
	thisPeriodStart, isEnd := dt.isEnd(tNow)

	// first SetValue invocation: initialize periodStart
	if isFirst {
		dt.periodStart = thisPeriodStart
		return // first period started return
	}

	// check for end of a period
	if !isEnd {
		if !dt.isFullPeriod {
			// not at end of a period and current period is the first, partial, period
			return // in the first, non-full, period return: no averaging data updates
		}
		// not at end of a full period
	} else {
		// beyond the end of the recording period: handle period change
		dt.updatePeriod(thisPeriodStart)
	}

	// update aggregate
	dt.nPeriod++
	dt.aggregate += float64(value)

	return
}

func (dt *Datapoint) CloneDatapoint() (datapoint parl.Datapoint) {
	return dt.cloneDatapoint(false)
}
func (dt *Datapoint) CloneDatapointReset() (datapoint parl.Datapoint) {
	return dt.cloneDatapoint(true)
}
func (dt *Datapoint) cloneDatapoint(reset bool) (datapoint parl.Datapoint) {
	tNow := time.Now()
	dt.lock.Lock()
	defer dt.lock.Unlock()

	// handle period end
	if thisPeriodStart, isEnd := dt.isEnd(tNow); isEnd {
		dt.updatePeriod(thisPeriodStart)
	}

	datapoint = &Datapoint{
		period:       dt.period,
		value:        dt.value,
		max:          dt.max,
		min:          dt.min,
		periodStart:  dt.periodStart,
		isFullPeriod: dt.isFullPeriod,
		nPeriod:      dt.nPeriod,
		aggregate:    dt.aggregate,
		average:      dt.average,
		n:            dt.n,
	}

	if !reset {
		return
	}

	dt.value = 0
	dt.max = 0
	dt.min = 0
	dt.periodStart = time.Time{}
	dt.isFullPeriod = false
	dt.nPeriod = 0
	dt.aggregate = 0
	dt.average = 0
	dt.n = 0

	return
}

func (dt *Datapoint) GetDatapoint() (value, max, min uint64, isValid bool, average float64, n uint64) {
	tNow := time.Now() // best possible tNow
	dt.lock.Lock()
	defer dt.lock.Unlock()

	// handle period end
	if thisPeriodStart, isEnd := dt.isEnd(tNow); isEnd {
		dt.updatePeriod(thisPeriodStart)
	}

	value = dt.value
	max = dt.max
	min = dt.min
	isValid = !dt.periodStart.IsZero()
	average = dt.average
	n = dt.n

	return
}

func (dt *Datapoint) DatapointValue() (value uint64) {
	value, _, _, _, _, _ = dt.GetDatapoint()
	return
}
func (dt *Datapoint) DatapointMax() (max uint64) {
	_, max, _, _, _, _ = dt.GetDatapoint()
	return
}
func (dt *Datapoint) DatapointMin() (min uint64) {
	_, _, min, _, _, _ = dt.GetDatapoint()
	return
}

// isEnd determines if the current period has ended.
// isEnd returns the start of the current period.
func (dt *Datapoint) isEnd(tNow time.Time) (thisPeriodStart time.Time, isEnd bool) {
	thisPeriodStart = tNow.Truncate(dt.period)

	// if the first SetValue has not happened, it cannot be isEnd
	if dt.periodStart.IsZero() {
		return
	}

	isEnd = !dt.periodStart.Equal(thisPeriodStart)
	return
}

// updatePeriod carries otu a period change.
// updatePeriod must only be executed when isEnd returns true.
func (dt *Datapoint) updatePeriod(thisPeriodStart time.Time) {
	if !dt.isFullPeriod {
		// the ending period was not a full period
		// no avergaing was done for that non-full period
		dt.isFullPeriod = true // first full period return
		return
	}

	// save valid average
	if dt.nPeriod > 0 {
		dt.average = dt.aggregate / float64(dt.nPeriod)
		dt.n = dt.nPeriod
	}

	// restart period averaging
	dt.periodStart = thisPeriodStart
	dt.aggregate = 0
	dt.nPeriod = 0
}
