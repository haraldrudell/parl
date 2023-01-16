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
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Counters is a container for counters, rate-counters and datapoints. Thread-Safe.
//   - a counter is Inc-Dec with: value running max
//   - a rate-counter is a counter with addtional measuring over short time periods:
//   - — value: rate of increase: current/max/average
//   - — running: rate up or down, max increase/decrease rate,
type Counters struct {
	lock    sync.RWMutex
	ordered []parl.CounterID       // behind lock: ordered list of counters and datapoints
	m       map[parl.CounterID]any // behind lock
	RateRunner
}

var _ parl.CounterSet = &Counters{}     // Counters is a parl.CounterSet
var _ parl.CounterSetData = &Counters{} // Counters is a parl.CounterSet

func newCounters(g0 parl.GoGen) (counters parl.Counters) {
	return &Counters{
		m:          map[parl.CounterID]any{},
		RateRunner: *NewRateRunner(g0),
	}
}

func (cs *Counters) GetOrCreateCounter(name parl.CounterID, period ...time.Duration) (counter parl.Counter) {
	counter = cs.getOrCreate(false, name, period...).(parl.Counter)
	return
}

func (cs *Counters) GetOrCreateDatapoint(name parl.CounterID, period time.Duration) (datapoint parl.Datapoint) {
	datapoint = cs.getOrCreate(true, name, period).(parl.Datapoint)
	return
}

func (cs *Counters) GetCounters() (list []parl.CounterID, m map[parl.CounterID]any) {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	list = slices.Clone(cs.ordered)
	m = maps.Clone(cs.m)

	return
}

func (cs *Counters) ResetCounters(stopRateCounters bool) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	_, m := cs.GetCounters()
	for _, item := range m {
		if counter, ok := item.(parl.CounterValues); ok {
			counter.CloneReset(stopRateCounters)
		} else if datapoint, ok := item.(parl.DatapointValue); ok {
			datapoint.CloneDatapointReset()
		} else {
			panic(perrors.ErrorfPF("Bad item in map: %T", item))
		}
	}
}

func (cs *Counters) Exists(name parl.CounterID) (exists bool) {
	exists = cs.getCounter(name) != nil
	return
}

func (cs *Counters) Value(name parl.CounterID) (value uint64) {
	if counter, ok := cs.getCounter(name).(interface{ Value() (value uint64) }); ok {
		value = counter.Value()
	}
	return
}

func (cs *Counters) Running(name parl.CounterID) (running uint64) {
	if counter, ok := cs.getCounter(name).(interface{ Running() (running uint64) }); ok {
		running = counter.Running()
	}
	return
}

func (cs *Counters) Max(name parl.CounterID) (max uint64) {
	if counter, ok := cs.getCounter(name).(interface{ Max() (max uint64) }); ok {
		max = counter.Max()
	}
	return
}

func (cs *Counters) Rates(name parl.CounterID) (rates map[parl.RateType]int64) {
	if counter, ok := cs.getCounter(name).(interface {
		Rates() (rates map[parl.RateType]int64)
	}); ok {
		rates = counter.Rates()
	}
	return
}

func (cs *Counters) DatapointValue(name parl.CounterID) (datapointValue uint64) {
	if counter, ok := cs.getCounter(name).(interface {
		DatapointValue() (datapointValue uint64)
	}); ok {
		datapointValue = counter.DatapointValue()
	}
	return
}

func (cs *Counters) DatapointMax(name parl.CounterID) (datapointMax uint64) {
	if counter, ok := cs.getCounter(name).(interface {
		DatapointMax() (datapointMax uint64)
	}); ok {
		datapointMax = counter.DatapointMax()
	}
	return
}

func (cs *Counters) DatapointMin(name parl.CounterID) (datapointMin uint64) {
	if counter, ok := cs.getCounter(name).(interface {
		DatapointMin() (datapointMin uint64)
	}); ok {
		datapointMin = counter.DatapointMin()
	}
	return
}

func (cs *Counters) GetDatapoint(name parl.CounterID) (value, max, min uint64, isValid bool, average float64, n uint64) {
	if counter, ok := cs.getCounter(name).(interface {
		GetDatapoint() (value, max, min uint64, isValid bool, average float64, n uint64)
	}); ok {
		value, max, min, isValid, average, n = counter.GetDatapoint()
	}
	return
}

func (cs *Counters) getOrCreate(isDatapoint bool, name parl.CounterID, period ...time.Duration) (item any) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	// check for existing counter or datapoint
	var ok bool
	if item, ok = cs.m[name]; ok {
		return // counter exists return
	}

	// instantiate counter or datapoint
	var period0 time.Duration
	if len(period) > 0 {
		period0 = period[0]
	}
	if !isDatapoint {
		if period0 == 0 {
			item = newCounter() // non-rate counter
		} else {
			item = newRateCounter(period0, cs)
		}
	} else {
		item = newDatapoint(period0)
	}

	// store the new counter or datapoint
	cs.ordered = append(cs.ordered, name)
	cs.m[name] = item

	return
}

func (cs *Counters) getCounter(name parl.CounterID) (counter any) {
	cs.lock.RLock()
	defer cs.lock.RUnlock()

	counter = cs.m[name]
	return
}
