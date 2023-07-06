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

// Averager is a container for averaging providing perioding and history.
//   - maxCount is the maximum interval over which average is calculated
type Averager[T constraints.Integer] struct {
	period   Period
	maxCount int

	sliceLock sync.RWMutex
	// may be nil
	// if length >0 first and last intervals are not nil
	intervals []*AverageInterval
}

// NewAverager returns an object that calculates average over a number of interval periods.
func NewAverager[T constraints.Integer]() (averager *Averager[T]) {
	return &Averager[T]{
		period:   *NewPeriod(averagerDefaultPeriod),
		maxCount: averagerDefaultCount,
	}
}

// NewAverager2 returns an object that calculates average over a number of interval periods.
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
func (av *Averager[T]) Add(value T, t ...time.Time) {
	if interval := av.getCurrent(av.period.Index(t...)); interval != nil {
		interval.Add(float64(value))
	}
}

// Average returns the average
//   - t is time, default time.Now()
//   - if sample count is zero, average is zero
//   - count is number of values making up the average
func (av *Averager[T]) Average(t ...time.Time) (average float64, count uint64) {
	av.getCurrent(av.period.Index(t...)) // ensure periods are updated
	av.sliceLock.RLock()
	defer av.sliceLock.RUnlock()

	for _, aPeriod := range av.intervals {
		if aPeriod == nil {
			continue
		}
		aPeriod.Aggregate(&count, &average)
	}
	if count > 0 {
		average = average / float64(count)
	}

	return
}

func (av *Averager[T]) Index(t ...time.Time) (index PeriodIndex) {
	return av.period.Index(t...)
}

// getCurrent ensures an interval for the current period exists and returns it
//   - interval is nil if periodIndex is too far in the past
func (av *Averager[T]) getCurrent(periodIndex PeriodIndex) (interval *AverageInterval) {
	if interval = av.getLastInterval(); interval != nil && interval.index == periodIndex {
		return // last existing period is the right one return
	}

	// add period for periodIndexNow to av.periods
	av.sliceLock.Lock()
	defer av.sliceLock.Unlock()

	// get last known period
	lastPeriod := periodIndex
	if interval != nil && interval.index > lastPeriod {
		lastPeriod = interval.index
	}

	// remove obsolete periods at start
	firstInterval := av.period.Sub(lastPeriod+1, av.maxCount)
	length := len(av.intervals)
	if length > 0 {
		var firstIndex PeriodIndex
		if firstIndex = av.intervals[0].index; firstIndex < firstInterval {
			pslices.TrimLeft(&av.intervals, av.period.Since(firstInterval, firstIndex))
			length = len(av.intervals)
		}
	}

	// ensure enough intervals
	numberOfIntervals := av.period.Available(periodIndex, av.maxCount)
	if length < numberOfIntervals {
		pslices.SetLength(&av.intervals, numberOfIntervals)
		length = numberOfIntervals
	}

	// ensure first interval exists
	if av.intervals[0] == nil {
		av.intervals[0] = NewAverageInterval(firstInterval)
	}

	// ensure last interval exists
	lastIndex := length - 1
	if interval = av.intervals[lastIndex]; interval == nil {
		interval = NewAverageInterval(periodIndex)
		av.intervals[lastIndex] = interval
	}

	if periodIndex < firstInterval {
		interval = nil
		return // periodIndex too far in the past return
	}

	interval = av.intervals[periodIndex-firstInterval]

	return
}

// getLastInterval returns the last period or nil if no period exists
func (av *Averager[T]) getLastInterval() (interval *AverageInterval) {
	av.sliceLock.RLock()
	defer av.sliceLock.RUnlock()

	if length := len(av.intervals); length > 0 {
		interval = av.intervals[length-1]
	}

	return
}
