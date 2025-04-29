/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
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
	lock parl.RWMutex
	// behind lock: ordered list of counters and datapoints
	ordered []parl.CounterID
	// behind lock
	m map[parl.CounterID]any
	RateRunner
}

// Counters is a parl.CounterSet
var _ parl.CounterSet = &Counters{}

// Counters is a parl.CounterSet
var _ parl.CounterSetData = &Counters{}

// Counters is a parl.CounterStore
var _ parl.CounterStore = &Counters{}

// newCounters returns a counter container.
func newCounters(g parl.GoGen) (counters parl.Counters) {
	var cImpl = &Counters{
		m: map[parl.CounterID]any{},
	}
	NewRateRunner(g, &cImpl.RateRunner)
	counters = cImpl
	return
}

func (c *Counters) GetOrCreateCounter(name parl.CounterID, period ...time.Duration) (counter parl.Counter) {
	counter = c.getOrCreate(false, name, period...).(parl.Counter)
	return
}

func (c *Counters) GetCounter(name parl.CounterID) (counter parl.Counter) {
	counter = c.getOrCreate(false, name).(parl.Counter)
	return
}

func (c *Counters) GetOrCreateDatapoint(name parl.CounterID, period time.Duration) (datapoint parl.Datapoint) {
	datapoint = c.getOrCreate(true, name, period).(parl.Datapoint)
	return
}

func (c *Counters) GetCounters() (list []parl.CounterID, m map[parl.CounterID]any) {
	defer c.lock.RLock().RUnlock()

	list = slices.Clone(c.ordered)
	m = maps.Clone(c.m)

	return
}

func (c *Counters) ResetCounters(stopRateCounters bool) {
	defer c.lock.Lock().Unlock()

	_, m := c.GetCounters()
	for _, item := range m {
		if counter, ok := item.(parl.CounterValues); ok {
			counter.GetReset()
		} else if datapoint, ok := item.(parl.DatapointValue); ok {
			datapoint.CloneDatapointReset()
		} else {
			panic(perrors.ErrorfPF("Bad item in map: %T", item))
		}
	}
}

func (c *Counters) Exists(name parl.CounterID) (exists bool) {
	exists = c.GetNamedCounter(name) != nil
	return
}

func (c *Counters) Value(name parl.CounterID) (value uint64) {
	if counter, ok := c.GetNamedCounter(name).(interface{ Value() (value uint64) }); ok {
		value = counter.Value()
	}
	return
}

func (c *Counters) Get(name parl.CounterID) (value, running, max uint64) {
	if counter, ok := c.GetNamedCounter(name).(interface {
		Get() (value, running, max uint64)
	}); ok {
		value, running, max = counter.Get()
	}
	return
}

func (c *Counters) Rates(name parl.CounterID) (rates map[parl.RateType]float64) {
	if counter, ok := c.GetNamedCounter(name).(interface {
		Rates() (rates map[parl.RateType]float64)
	}); ok {
		rates = counter.Rates()
	}
	return
}

func (c *Counters) DatapointValue(name parl.CounterID) (datapointValue uint64) {
	if counter, ok := c.GetNamedCounter(name).(interface {
		DatapointValue() (datapointValue uint64)
	}); ok {
		datapointValue = counter.DatapointValue()
	}
	return
}

func (c *Counters) DatapointMax(name parl.CounterID) (datapointMax uint64) {
	if counter, ok := c.GetNamedCounter(name).(interface {
		DatapointMax() (datapointMax uint64)
	}); ok {
		datapointMax = counter.DatapointMax()
	}
	return
}

func (c *Counters) DatapointMin(name parl.CounterID) (datapointMin uint64) {
	if counter, ok := c.GetNamedCounter(name).(interface {
		DatapointMin() (datapointMin uint64)
	}); ok {
		datapointMin = counter.DatapointMin()
	}
	return
}

func (c *Counters) GetDatapoint(name parl.CounterID) (value, max, min uint64, isValid bool, average float64, n uint64) {
	if counter, ok := c.GetNamedCounter(name).(interface {
		GetDatapoint() (value, max, min uint64, isValid bool, average float64, n uint64)
	}); ok {
		value, max, min, isValid, average, n = counter.GetDatapoint()
	}
	return
}

func (c *Counters) GetNamedCounter(name parl.CounterID) (counter any) {
	defer c.lock.RLock().RUnlock()

	counter = c.m[name]
	return
}

func (c *Counters) getOrCreate(isDatapoint bool, name parl.CounterID, period ...time.Duration) (item any) {
	defer c.lock.Lock().Unlock()

	// check for existing counter or datapoint
	var ok bool
	if item, ok = c.m[name]; ok {
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
			var r = newRateCounter()
			item = r
			c.AddTask(period0, r)
		}
	} else {
		item = newDatapoint(period0)
	}

	// store the new counter or datapoint
	c.ordered = append(c.ordered, name)
	c.m[name] = item

	return
}
