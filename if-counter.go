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
	Rates(name CounterID) (rates map[RateType]int64)
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

// Counter is the data provider interface for a counter with Inc Dec SetValue operations.
//   - value is a monotonously increasing value counting Inc and positive Add occurrences.
//     value is used to count the total occurrences of something.
//   - running is the sum of Inc Dec and Add operations, but alaways 0 or greater.
//     running is used to count the currently operating instances of something.
//   - max is the maxmimum value reached by running
//
// Counter is thread-safe.
type Counter interface {
	Inc() (counter Counter) // Inc increments the counter. Supports method chaining
	Dec() (counter Counter) // Dec decrements the counter but not beyond zero. Supports method chaining
	// Add adds a positive or negative delta
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

// CounterValues is the consumer interface for a counter.
//   - value holds the current value from Inc or SetValue operations
//   - running hold the combining result of Inc and Dec operations. It is not affected by SetValue
//   - max is the maximum of running counter or SetValue operations
//   - values are uint64, Counter is thread-safe
type CounterValues interface {
	Clone() (counterValues CounterValues)                           // Clone takes a snapshot of a counter state.
	CloneReset(stopRateCounters bool) (counterValues CounterValues) // CloneReset takes a snapshot of a counter state and resets it to its initial state.
	Get() (value, running, max uint64)
	Value() (value uint64)
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
	CounterValues // Clone() CloneReset() Get() Value() Running() Max() DidSetValue()
	Rates() (rates map[RateType]int64)
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
	NewCounters(useCounters bool, g0 GoGen) (counters Counters)
}

const (
	ValueRate        RateType = iota // current rate of increase in value
	ValueMaxRate                     // max seen rate of increase in value
	ValueRateAverage                 // average rate of increase in value during last 10 periods
	RunningRate
	RunningMaxRate
	RunningMaxDecRate
	RunningAverage
	NotAValue // NotAValue is an internal stand-in value indicating a value not in use
)

type RateType int
