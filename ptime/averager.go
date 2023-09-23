/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

const (
	averagerDefaultPeriod = time.Second
	averagerDefaultCount  = 10
)

// Averager is a container for averaging integer values providing perioding and history.
//   - maxCount is the maximum interval over which average is calculated
//   - average is the average value of provided datapoints during all periods
//   - because averaging is over 10 or many periods, the average will change slowly
type Averager[T constraints.Integer] struct {
	// period provides linear interval-numbering from a time.Time timestamp
	period   Period
	maxCount int

	sliceLock sync.RWMutex
	// each pointer has a period index, datapoint count and datapoint aggregate
	//	- may be nil
	//	- if no value was provided during a particular period, that pointer is nil
	//	- if length >0, first and last intervals are not nil
	//	- thread-safe behind sliceLock
	intervals []*AverageInterval
}

// NewAverager returns an object that calculates average over a number of interval periods.
//   - interval-length is 1 s
//   - averaging over 10 intervals
func NewAverager[T constraints.Integer]() (averager *Averager[T]) {
	return &Averager[T]{
		period:   *NewPeriod(averagerDefaultPeriod),
		maxCount: averagerDefaultCount,
	}
}

// NewAverager2 returns an object that calculates average over a number of interval periods.
//   - period 0 means default 1 s interval-length
//   - periodCount 0 means averaging over 10 intervals
func NewAverager2[T constraints.Integer](period time.Duration, periodCount int) (averager *Averager[T]) {
	if period == 0 {
		period = averagerDefaultPeriod
	} else if period < 1 {
		perrors.ErrorfPF("period cannot be less than 1 ns: %s", period)
	}
	if periodCount == 0 {
		periodCount = averagerDefaultCount
	} else if periodCount < 2 {
		perrors.ErrorfPF("periodCount cannot be less than 2: %d", periodCount)
	}
	return &Averager[T]{
		period:   *NewPeriod(period),
		maxCount: periodCount,
	}
}

// Add adds a new value basis for averaging
//   - value is the sample value
//   - t is the sample time, default time.Now()
func (a *Averager[T]) Add(value T, t ...time.Time) {
	if interval := a.getCurrent(a.period.Index(t...)); interval != nil {
		interval.Add(float64(value))
	}
}

// Average returns the average
//   - t is time, default time.Now()
//   - if sample count aggregated across the slice zero, average is zero
//   - count is number of values making up the average
func (a *Averager[T]) Average(t ...time.Time) (average float64, count uint64) {

	// invoking getCurrent ensure periods are updated
	//	- if no recent datapoints were provided, the slice may contain old values
	a.getCurrent(a.period.Index(t...))
	a.sliceLock.RLock()
	defer a.sliceLock.RUnlock()

	// calculate the average
	for _, aPeriod := range a.intervals {
		if aPeriod == nil {
			continue // nil average pointers is allowed
		}
		aPeriod.Aggregate(&count, &average)
	}
	if count > 0 {
		average = average / float64(count)
	}

	return
}

// TODO 230726 Delete
// func (a *Averager[T]) Index(t ...time.Time) (index PeriodIndex) {
// 	return a.period.Index(t...)
// }

// getCurrent ensures an interval for the current period exists and returns it
//   - interval is nil if periodIndex is too far in the past
//   - the slice is updated to include periodIndex if it is after the last entry
func (a *Averager[T]) getCurrent(periodIndex PeriodIndex) (interval *AverageInterval) {
	if interval = a.getLastInterval(); interval != nil && interval.index == periodIndex {
		return // last existing period is the right one return
	}

	// ensure the slice is up-to-date and
	// return the requested period if it is not out of range
	a.sliceLock.Lock()
	defer a.sliceLock.Unlock()

	// 1. determine at what period the slice should end

	// the latest known period
	//	- either the requested value or
	//	- the last existing in slice
	var lastPeriod = periodIndex
	if interval != nil && interval.index > lastPeriod {
		lastPeriod = interval.index
	}
	// firstInterval is the lowest allowed period number considering:
	//	- maxCount and period0
	var firstInterval = a.period.Sub(lastPeriod+1, a.maxCount)
	// length is the number of entries in the slice
	var length = len(a.intervals)

	// 2. remove any stale slice entries at start of slice
	if length > 0 {
		var firstIndex PeriodIndex
		if firstIndex = a.intervals[0].Index(); firstIndex < firstInterval {
			pslices.TrimLeft(&a.intervals, a.period.Since(firstInterval, firstIndex))
			length = len(a.intervals)
		}
	}
	// numberOfIntervals is the length the slice should have to include periodIndex
	var numberOfIntervals = a.period.Available(periodIndex, a.maxCount)

	// 3. extend the slice as necessary
	if length < numberOfIntervals {
		pslices.SetLength(&a.intervals, numberOfIntervals)
		length = numberOfIntervals
	}

	// ensure first interval exists
	if a.intervals[0] == nil {
		a.intervals[0] = NewAverageInterval(firstInterval)
	}

	// ensure last interval exists
	var lastIndex = length - 1
	if interval = a.intervals[lastIndex]; interval == nil {
		interval = NewAverageInterval(periodIndex)
		a.intervals[lastIndex] = interval
	}

	if periodIndex < firstInterval {
		interval = nil
		return // periodIndex too far in the past return
	}

	interval = a.intervals[periodIndex-firstInterval]

	return
}

// getLastInterval returns the last period or nil if no period exists
//   - the last period provide the current end of the value series
//   - if no recent datapoints were provided,
//     the last period may be stale
func (a *Averager[T]) getLastInterval() (interval *AverageInterval) {
	a.sliceLock.RLock()
	defer a.sliceLock.RUnlock()

	if length := len(a.intervals); length > 0 {
		interval = a.intervals[length-1]
	}

	return
}
