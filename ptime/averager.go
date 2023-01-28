/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

const (
	averagerDefaultPeriod = time.Second
	averagerDefaultCount  = 10
)

// Averager is a container for averaging providing perioding and history.
//   - Period provides uint64 zero-based numbered period of fixed interval length
//   - Period provides first period index and fractional usage of the first period
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

func (av *Averager[T]) Add(value T, t ...time.Time) {
	av.getCurrent(av.period.Index(t...)).Add(float64(value))
}

func (av *Averager[T]) Average(t ...time.Time) (average float64) {
	var count uint64
	av.getCurrent(av.period.Index(t...)) // ensure peiods are updated
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

// getCurrent ensures an interval for the current period exists and returns it
func (av *Averager[T]) getCurrent(periodIndexNow PeriodIndex) (interval *AverageInterval) {
	if interval = av.getLastInterval(); interval != nil && interval.index == periodIndexNow {
		return // last existing period is the right one return
	}

	// add period for periodIndexNow to av.periods
	av.sliceLock.Lock()
	defer av.sliceLock.Unlock()

	// remove obsolete periods at start
	firstInterval := av.period.Sub(periodIndexNow, av.maxCount)
	length := len(av.intervals)
	if length > 0 {
		var firstIndex PeriodIndex
		if firstIndex = av.intervals[0].index; firstIndex < firstInterval {
			pslices.TrimLeft(&av.intervals, av.period.Since(firstInterval, firstIndex))
			length = len(av.intervals)
		}
	}

	// ensure enough intervals
	numberOfIntervals := av.period.Available(periodIndexNow, av.maxCount)
	if length < numberOfIntervals {
		pslices.SetLength(&av.intervals, numberOfIntervals)
		length = numberOfIntervals
	}

	// ensure first interval exists
	if av.intervals[0] == nil {
		av.intervals[0] = NewAverageInterval(firstInterval)
	}

	// ensure current interval exists
	lastIndex := length - 1
	if interval = av.intervals[lastIndex]; interval == nil {
		interval = NewAverageInterval(periodIndexNow)
		av.intervals[lastIndex] = interval
	}

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
