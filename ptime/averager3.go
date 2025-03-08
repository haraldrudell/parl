/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"golang.org/x/exp/constraints"
)

// Averager3 is an average container with last, average and max values.
type Averager3[T constraints.Integer] struct {
	Averager[T]
	max   cyclebreaker.AtomicMax[T]
	index cyclebreaker.AtomicMax[Epoch]
	last  time.Duration // atomic
}

// NewAverager3 returns an calculator for last, average and max values.
func NewAverager3[T constraints.Integer]() (averager *Averager3[T]) {
	averager = &Averager3[T]{}
	NewAverager(&averager.Averager)
	return
}

// Status returns "[last]/[average]/[max]"
//   - if no values at all: "-/-/-"
//   - if no values avaiable to calculate average "1s/-/10s"
func (av *Averager3[T]) Status() (s string) {

	// check if any values are present
	max, hasValue := av.max.Max()
	if !hasValue {
		s = "-/-/-"
		return // no values return
	}

	// check average
	average, count := av.Averager.Average()
	if count > 0 {
		s = Duration(time.Duration(average))
	} else {
		s = "-"
	}

	s = Duration(time.Duration(atomic.LoadInt64((*int64)(&av.last)))) + "/" +
		s + "/" +
		Duration(time.Duration(max))

	return
}

// Add updates last, average and max values
func (av *Averager3[T]) Add(value T, t ...time.Time) {
	av.max.Value(value)
	if av.index.Value(EpochNow(t...)) {
		atomic.StoreInt64((*int64)(&av.last), int64(value))
	}
	av.Averager.Add(value, t...)
}
