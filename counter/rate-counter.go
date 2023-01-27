/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// RateCounter is a value/running/max counter with averaging.
package counter

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl"
	"golang.org/x/exp/maps"
)

const (
	averagerSize = 10
)

// RateCounter is a value/running/max counter with averaging.
//   - rate of increase, maximum and average rate of increase in value
//   - rate of increase, maximum increase and decrease rates and average of value
type RateCounter struct {
	Counter // value-running-max atomic-access container
	period  Period

	lock       sync.Mutex
	hasValues  bool   // indicates that value and running was initialized at start of period
	value      uint64 // value at beginning of period
	running    uint64 // running at beginning of period
	m          map[parl.RateType]int64
	valueAvg   Averager
	runningAvg Averager
}

var _ parl.RateCounterValues = &RateCounter{} // RateCounter is parl.RateCounterValues

func newRateCounter(interval time.Duration, cs *Counters) (counter parl.Counter) {
	c := RateCounter{
		period: *NewPeriod(interval),
		m:      map[parl.RateType]int64{},
	}
	InitAverager(&c.valueAvg, averagerSize)
	InitAverager(&c.runningAvg, averagerSize)
	cs.AddTask(interval, &c)
	return &c
}

func (rc *RateCounter) Rates() (rates map[parl.RateType]int64) {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	rates = maps.Clone(rc.m)
	return
}
func (rc *RateCounter) Do() {
	rc.lock.Lock()
	defer rc.lock.Unlock()

	// get current values
	value, running, _ := rc.Counter.Get()
	// for running, average its actual value
	rc.m[parl.RunningAverage] = int64(rc.runningAvg.Add(running))
	if !rc.hasValues {

		// first invocation: initialize values
		rc.value = value
		rc.running = running
		rc.hasValues = true
		return // populated start of period return
	}

	// update rates
	rc.do(rc.value, value, parl.ValueRate, parl.ValueMaxRate, parl.NotAValue)
	rc.do(rc.running, running, parl.RunningRate, parl.RunningMaxRate, parl.RunningMaxDecRate)
	rc.value = value
	rc.running = running

	// for value, average its rate of increase
	rc.m[parl.ValueRateAverage] = int64(rc.valueAvg.Add(uint64(rc.m[parl.ValueRate])))
}

// do performs rate counter calculation over a period starting at from value and
// ending at to value.
func (rc *RateCounter) do(from, to uint64, rateX, maxRateX, maxDecRateX parl.RateType) {
	mapp := rc.m
	if to == from {
		return // value is zero, rate is zero return
	} else if to > from { // not negative
		rate := int64(to - from)
		mapp[rateX] = rate
		if rate > mapp[maxRateX] {
			mapp[maxRateX] = rate
		}
		return // positive rate return
	}

	if maxDecRateX == parl.NotAValue {
		return
	}
	rate := -int64(from - to)
	if rate < mapp[maxDecRateX] {
		mapp[maxDecRateX] = rate
	}
	// negative rate return
}
