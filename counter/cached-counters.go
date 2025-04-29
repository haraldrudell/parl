/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// CachedCounters reduces counter map contention by only looking up the counterID once
type CachedCounters struct {
	// uncached counters accessible by ID
	counterStore parl.CounterStore
	// cached regular counters
	counterValues map[parl.CounterID]parl.CounterValues
	// cached rate counters
	rateCounterValues map[parl.CounterID]parl.RateCounterValues
	// cached datapoints
	datapointValue map[parl.CounterID]parl.DatapointValue
}

// NewCachedCounters returns a cache of parl.CounterID reducing map contention
func NewCachedCounters(counterStore parl.CounterStore, fieldp ...*CachedCounters) (cachedCounters *CachedCounters) {

	// get cachedCounters
	if len(fieldp) > 0 {
		cachedCounters = fieldp[0]
	}
	if cachedCounters == nil {
		cachedCounters = &CachedCounters{}
	}

	*cachedCounters = CachedCounters{
		counterStore:      counterStore,
		counterValues:     make(map[parl.CounterID]parl.CounterValues),
		rateCounterValues: make(map[parl.CounterID]parl.RateCounterValues),
		datapointValue:    make(map[parl.CounterID]parl.DatapointValue),
	}
	return
}

// CounterValues never returns nil
//   - returns consumer interface for regular counter: Get and Value methods
func (c *CachedCounters) CounterValues(counterID parl.CounterID) (counterValues parl.CounterValues) {

	// try cached value
	if counterValues = c.counterValues[counterID]; counterValues != nil {
		return
	}

	// populate cache

	// counter is the provider interface of a regular counter,
	// what is returned by GetOrCreateCounter
	var counter = c.counterStore.GetCounter(counterID)
	counterValues = counter.(parl.CounterValues)
	c.counterValues[counterID] = counterValues

	return
}

// RateCounter may return nil
//   - returns consumer interface for rate counter: Get Value Rates methods
func (c *CachedCounters) RateCounter(counterID parl.CounterID) (rateCounterValues parl.RateCounterValues) {

	// try cached value
	if rateCounterValues = c.rateCounterValues[counterID]; rateCounterValues != nil {
		return
	}

	// populate cache
	var counterAny any
	var ok bool
	if counterAny = c.counterStore.GetNamedCounter(counterID); counterAny == nil {
		return
	} else if rateCounterValues, ok = counterAny.(parl.RateCounterValues); !ok {
		panic(perrors.ErrorfPF("not a rate counter: counter ID: %s type: %T", counterID, counterAny))
	}
	c.rateCounterValues[counterID] = rateCounterValues

	return
}

// RateCounter may return nil
//   - returns consumer interface for rate counter: Get Value Rates methods
func (c *CachedCounters) DataPoint(counterID parl.CounterID) (datapointValue parl.DatapointValue) {

	// try cached value
	if datapointValue = c.datapointValue[counterID]; datapointValue != nil {
		return
	}

	// populate cache
	var counterAny any
	var ok bool
	if counterAny = c.counterStore.GetNamedCounter(counterID); counterAny == nil {
		return
	} else if datapointValue, ok = counterAny.(parl.DatapointValue); !ok {
		panic(perrors.ErrorfPF("not a datapoint: counter ID: %s type: %t", counterID, counterAny))
	}
	c.datapointValue[counterID] = datapointValue

	return
}
