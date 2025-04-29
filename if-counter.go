/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"time"
)

// CounterID is a unique type containing counter names
type CounterID string

// data provider side

// Counters is the data provider interface for a counter set.
//   - max and running values are offered
//   - Counters and datapointy are thread-safe
//   - counters may be used to determine that code abide
//     by intended paralellism and identifying hangs or abnormalisms.
//   - Printing counters every second can verify adequate progress and
//     possibly identify blocking of threads or swapping and garbage collection
//     outages.
type Counters interface {
	// GetOrCreateCounter is used by the data provider of a counter.
	//	- A counter has Inc and Dec operations.
	//	- Counters can be used to track number of currently operating invocations.
	//	- if period is present and positive, the counter created is a RateCounter.
	// Rate counters use threads, one per distinct period requested.
	// Rate counter threads are monitored by the provided Go group and failure
	// may force a counter restart or cause application termination. Failures
	// are typically caused by panics in the counter update task.
	//	- Counter is the data-provider side of a counter
	//	- CounterValues is the consumer side of a counter
	GetOrCreateCounter(name CounterID, period ...time.Duration) (counter Counter)
	// GetOrCreateCounter is used by the data provider of a datapoint.
	// A datapoint supports SetValue operation.
	// A datapoint tracks a quantity such as a latency value.
	GetOrCreateDatapoint(name CounterID, period time.Duration) (datapoint Datapoint)
}

// CounterSetData provides simple access to a set of counters, rate counters and datapoints/
type CounterSetData interface {
	Exists(name CounterID) (exists bool)
	// Value returns the monotonically increasing value for a possible plain counter
	//	- if no such counter exists, 0
	Value(name CounterID) (value uint64)
	// Running returns the increasing and decreasing running value for a possible plain counter
	//	- if no such counter exists, 0
	Get(name CounterID) (value, running, max uint64)
	// Rates returns the rate values a possible rate counter
	//	- if no such counter exists or values are not yet available, nil
	Rates(name CounterID) (rates map[RateType]float64)
	// DatapointValue returns the latest value a possible datapoint
	//	- if no such datapoint exists, 0
	DatapointValue(name CounterID) (value uint64)
	// DatapointMax returns the highest seen value for a possible datapoint
	//	- if no such datapoint exists, 0
	DatapointMax(name CounterID) (max uint64)
	// DatapointMin returns the lowest seen value for a possible datapoint
	//	- if no such datapoint exists, 0
	DatapointMin(name CounterID) (min uint64)
	// GetDatapoint returns dfatapoint data for a possible datapoint
	//	- if no such datapoint exists, 0
	GetDatapoint(name CounterID) (value, max, min uint64, isValid bool, average float64, n uint64)
}

// Counter is the data provider interface for a counter
//   - Inc Dec Add operations, Thread-safe
type Counter interface {
	// Inc increments the counter. Thread-Safe, method chaining
	Inc() (counter Counter)
	// Dec decrements the counter but not below zero. Thread-Safe, method chaining
	Dec() (counter Counter)
	// Add adds a positive or negative delta. Thread-Safe, method chaining
	Add(delta int64) (counter Counter)
}

// Datapoint tracks a value with average, max-min, and increase/decrease rates.
type Datapoint interface {
	// SetValue records a value at time.Now().
	// SetValue supports method chaining.
	SetValue(value uint64) (datapoint Datapoint)
}

// consumer side

// CounterSet is the consumer interface for a counter set
type CounterSet interface {
	// GetCounters gets a list and a map for consuming counter data
	GetCounters() (orderedKeys []CounterID, m map[CounterID]any)
	// ResetCounters resets all counters to their initial state
	ResetCounters(stopRateCounters bool)
}

// CounterStore is a CounterSet consumer interface facilitating caching
type CounterStore interface {
	// GetOrCreateCounter retrieves a regular counter
	//	- never returns nil
	//	- type asserted to CounterValues
	GetCounter(name CounterID) (counter Counter)
	//GetCounter retrieves a counter that must exist
	//	- may return nil
	//	- type asserted to RateCounterValues or DatapointValue
	GetNamedCounter(name CounterID) (counter any)
}

// CounterValues is the consumer interface for a counter. Thread-safe
type CounterValues interface {
	// Get returns value/running/max with integrity. Thread-Safe
	//	- value is the monotonically increasing value
	//	- running is the fluctuating running value
	//	- max is the highest value running has had
	Get() (value, running, max uint64)
	// GetReset returns value/running/max with integrity and resets the counter. Thread-Safe
	//	- value is the monotonically increasing value
	//	- running is the fluctuating running value
	//	- max is the highest value running has had
	GetReset() (value, running, max uint64)
	// Value returns the monotonically increasing value. Thread-Safe
	//	- number of Inc invocations and positive Adds
	Value() (value uint64)
	// Running returns the fluctuating running value. Thread-Safe
	//	- number of Inc less Dec invocations and sum of Adds
	//	- never below 0
	Running() (running uint64)
	// Max returns the highest value running has had. Thread-Safe
	Max() (max uint64)
}

// CounterValues is the consumer interface for a rate counter.
// The rate counter provides rates over a provided time period in a map of int64 values.
//   - ValueRate current rate of increase in value
//   - ValueMaxRate the maxmimum seen rate of increase in value
//   - ValueRateAverage the average rate of increase in value taken over up to 10 periods
//   - RunningRate the current rate of change in running, may be negative
//   - RunningMaxRate the max positive rate of increase seen in running
//   - RunningMaxDecRate the max rate of decrease in running, a 0 or negative value
//   - RunningAverage the average of running taken over up to 10 periods
type RateCounterValues interface {
	CounterValues // Get() GetReset() Value() Running() Max()
	Rates() (rates map[RateType]float64)
}

// Rate describes a rate datapoint.
//   - may be positive or negative
type Rate interface {
	Clone() (rate Rate)
	Delta() (delta uint64)
	Duration() (duration time.Duration)
	HasValue() (hasValue bool)
	fmt.Stringer
}

type DatapointValue interface {
	CloneDatapoint() (datapoint Datapoint)      // Clone takes a snapshot of a counter state.
	CloneDatapointReset() (datapoint Datapoint) // CloneReset takes a snapshot of a counter state and resets it to its initial state.
	GetDatapoint() (value, max, min uint64, isValid bool, average float64, n uint64)
	DatapointValue() (value uint64)
	DatapointMax() (max uint64)
	DatapointMin() (min uint64)
}

// CountersFactory is an abstract counter factory.
// CountersFactory enables providing of different counter implementations.
type CountersFactory interface {
	// NewCounters returns a counter container.
	// is useCounters is false, the container does not actually do any counting.
	NewCounters(useCounters bool, g GoGen) (counters Counters)
}

const (
	// current rate of increase in value
	ValueRate RateType = iota
	// max seen rate of increase in value
	ValueMaxRate
	// average rate of increase in value during last 10 periods
	ValueRateAverage
	RunningRate
	RunningMaxRate
	RunningMaxDecRate
	// accumulated change in running over several intervals
	//	- running value goes up and down
	RunningAverage
	// NotAValue is an internal stand-in value indicating a value not in use
	NotAValue
)

type RateType int

const (
	// [counter.CountersFactoryNewCounters] counters are active
	CountersYes CounterType = iota + 1
	// [counter.CountersFactoryNewCounters] counters do nothing
	CountersNo
)

type CounterType uint8
