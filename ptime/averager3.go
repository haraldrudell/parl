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

type Averager3[T constraints.Integer] struct {
	Averager[T]
	max  cyclebreaker.AtomicMax[T]
	last time.Duration // atomic
}

// NewAverager3 returns an object that calculates average over a number of interval periods.
func NewAverager3[T constraints.Integer]() (averager *Averager3[T]) {
	return &Averager3[T]{Averager: *NewAverager[T]()}
}

func (av *Averager3[T]) Status() (s string) {
	max, hasValue := av.max.Max()
	if !hasValue {
		s = "-/-/-"
		return
	}
	s = Duration(time.Duration(atomic.LoadInt64((*int64)(&av.last)))) + "/" +
		Duration(time.Duration(av.Averager.Average())) + "/" +
		Duration(time.Duration(max))
	return
}

func (av *Averager3[T]) Add(value T, t ...time.Time) {
	av.max.Value(value)
	atomic.StoreInt64((*int64)(&av.last), int64(value))
	av.Averager.Add(value, t...)
}
