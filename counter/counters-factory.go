/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"time"

	"github.com/haraldrudell/parl"
)

var CountersFactory parl.CountersFactory = &countersFactory{}

type countersFactory struct{}

// useCounters true:
func (f *countersFactory) NewCounters(useCounters bool, g parl.GoGen) (counters parl.Counters) {
	if !useCounters {
		return &countersNil{}
	}
	return newCounters(g)
}

type countersNil struct{}

var _ parl.Counters = &countersNil{}

func (tn *countersNil) GetOrCreateCounter(name parl.CounterID, period ...time.Duration) (counter parl.Counter) {
	return &counterNil{}
}
func (tn *countersNil) GetOrCreateDatapoint(name parl.CounterID, period time.Duration) (datapoint parl.Datapoint) {
	return &datapointNil{}
}

type counterNil struct{}

var _ parl.Counter = &counterNil{}

//var _ parl.RateCounterValues = &counterNil{}

func (tn *counterNil) Inc() (counters parl.Counter)            { return tn }
func (tn *counterNil) Dec() (counters parl.Counter)            { return tn }
func (tn *counterNil) Add(delta int64) (counters parl.Counter) { return tn }

type datapointNil struct{}

var _ parl.Datapoint = &datapointNil{}

func (tn *datapointNil) SetValue(value uint64) (datapoint parl.Datapoint) { return }
