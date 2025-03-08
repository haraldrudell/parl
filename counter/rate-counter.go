/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"time"

	"github.com/haraldrudell/parl"
	"golang.org/x/exp/maps"
)

// RateCounter is a value/running/max counter with averaging.
//   - rate of increase, maximum and average rate of increase in value
//   - rate of increase, maximum increase and decrease rates and average of value
type RateCounter struct {
	// value-running-max atomic-access container
	Counter

	lock parl.Mutex
	// used to calculate true duration of each period
	lastDoInvocation time.Time
	// indicates that value and running was initialized at start of period
	hasValues bool
	// value at beginning of period
	value uint64
	// running at beginning of period
	running    uint64
	m          map[parl.RateType]float64
	valueAvg   Averager
	runningAvg Averager
}

// RateCounter is parl.RateCounterValues
var _ parl.RateCounterValues = &RateCounter{}

// newRateCounter returns a rate-counter, an extension to a regular 3-value counter
func newRateCounter() (counter *RateCounter) {
	counter = &RateCounter{
		lastDoInvocation: time.Now(),
		m:                make(map[parl.RateType]float64),
	}
	NewAverager(averagerSize, &counter.valueAvg)
	NewAverager(averagerSize, &counter.runningAvg)
	return
}

// Rates returns a map of current rate-results
func (r *RateCounter) Rates() (rates map[parl.RateType]float64) {
	defer r.lock.Lock().Unlock()

	rates = maps.Clone(r.m)
	return
}

// Do completes rate-calculations for a period
//   - Do is invoked by the task container
//   - at is an accurate timestamp, ie. not from a time.Interval
func (r *RateCounter) Do(at time.Time) {
	defer r.lock.Lock().Unlock()

	// calculate interval duration: positive, at least minInterval
	var duration = at.Sub(r.lastDoInvocation)
	if duration < minInterval {
		return // ignore if too close to last occasion, or negative
	}
	r.lastDoInvocation = at
	var seconds = duration.Seconds()

	// get current values from the underlying rate/running/max counter
	var value, running, _ = r.Counter.Get()

	// running goes up and down
	//	- running average is total change over the period divided by duration of all periods
	r.m[parl.RunningAverage] = r.runningAvg.Add(running, duration)

	// rates requires values from beginning of the period
	//	- collect values at beginning of the very first period
	if !r.hasValues {

		// first invocation: initialize values
		r.value = value
		r.running = running
		r.hasValues = true
		return // populated start of period return
	}

	// update rates since previous period
	r.do(r.value, value, seconds, parl.ValueRate, parl.ValueMaxRate, parl.NotAValue)
	r.do(r.running, running, seconds, parl.RunningRate, parl.RunningMaxRate, parl.RunningMaxDecRate)

	// update last period’s values
	r.value = value
	r.running = running

	// for value, average its rate of increase
	//	- since value is monotonically increasing, this is meaningful
	r.m[parl.ValueRateAverage] = r.valueAvg.Add(uint64(r.m[parl.ValueRate]), duration)
}

// do performs rate counter calculation over a period starting at from value and
// ending at to value.
//   - from is value at beginning of period
//   - to is current value
func (r *RateCounter) do(from, to uint64, seconds float64, rateIndex, maxRateIndex, maxDecRateIndex parl.RateType) {
	var m = r.m
	if to == from {
		return // value is zero, rate is zero return: keep last rate
	}

	// calculate positive rate and max rate
	if to > from { // not negative
		rate := float64(to-from) / seconds
		m[rateIndex] = rate
		if rate > m[maxRateIndex] {
			m[maxRateIndex] = rate
		}
		return // positive rate return
	}

	// calculate decreasing rate
	if maxDecRateIndex == parl.NotAValue {
		return // max decrease rate should not be calculated
	}
	var rate = (float64(to) - float64(from)) / seconds
	if rate < m[maxDecRateIndex] {
		m[maxDecRateIndex] = rate
	}
	// negative rate return
}

const (
	averagerSize = 10
	minInterval  = time.Microsecond
)
