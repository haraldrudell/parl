/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Counter is a counter without rate information.
package counter

import (
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

// Counter is a counter without rate information. Thread-safe.
//   - value: monotonically increasing: Inc
//   - running: Inc - Dec
//   - max: the highest value of running
//   - Counter implements parl.Counter and parl.CounterValues.
//
// Note: because Counter uses atomics and not lock, data integrity is not
// guaranteed or achievable. Methods Clone, Get or the order of
// Value then Running then Max ensures that data is valid and consistent but
// possibly of progressing revision.
type Counter struct {
	value   uint64 // atomic
	running uint64 // atomic
	max     uint64 // atomic
}

var _ parl.CounterValues = &Counter{} // Counter is parl.CounterValues

func newCounter() (counter parl.Counter) { // Counter is parl.Counter
	return &Counter{}
}

func (cn *Counter) Inc() (counter parl.Counter) {
	counter = cn

	// update value and running
	atomic.AddUint64(&cn.value, 1)
	cn.updateMax(atomic.AddUint64(&cn.running, 1)) // returns the new value

	return
}

func (cn *Counter) Dec() (counter parl.Counter) {
	counter = cn

	for {
		current := atomic.LoadUint64(&cn.running) // current value
		if current == 0 {
			break // do not decrement beyond 0
		}
		newCurrent := current - 1 // 0…
		if atomic.CompareAndSwapUint64(&cn.running, current, newCurrent) {
			break // decrement succeeded
		}
	}
	return
}

func (cn *Counter) Add(delta int64) (counter parl.Counter) {
	counter = cn

	if delta == 0 {
		return
	} else if delta > 0 {
		u64 := uint64(delta)
		atomic.AddUint64(&cn.value, u64)
		cn.updateMax(atomic.AddUint64(&cn.running, u64))
		return
	}

	// delta negative
	u64 := uint64(-delta)
	for {
		running := atomic.LoadUint64(&cn.running)
		var newValue uint64
		if u64 < running {
			newValue = running - u64
		}
		if atomic.CompareAndSwapUint64(&cn.running, running, newValue) {
			return
		}
	}
}

func (cn *Counter) updateMax(running uint64) {
	// ensure max is at least running
	for {
		max := atomic.LoadUint64(&cn.max) // get current max value
		if running <= max {
			return // no update is required

			// CompareAndSwapUint64 updates to running if value still matches max
		} else if atomic.CompareAndSwapUint64(&cn.max, max, running) {
			return // update was successful
		}
	}
}

func (cn *Counter) Clone() (counterValues parl.CounterValues) {
	cv := Counter{}
	counterValues = &cv

	cv.value = atomic.LoadUint64(&cn.value)
	cv.running = atomic.LoadUint64(&cn.running)
	cv.max = atomic.LoadUint64(&cn.max)
	return
}

func (cn *Counter) CloneReset(stopRateCounters bool) (counterValues parl.CounterValues) {
	counterValues = cn.Clone()

	atomic.StoreUint64(&cn.value, 0)
	atomic.StoreUint64(&cn.running, 0)
	atomic.StoreUint64(&cn.max, 0)
	return
}

func (cn *Counter) Get() (value uint64, running uint64, max uint64) {
	value = atomic.LoadUint64(&cn.value)
	running = atomic.LoadUint64(&cn.running)
	max = atomic.LoadUint64(&cn.max)
	return
}

func (cn *Counter) Value() (value uint64) {
	value = atomic.LoadUint64(&cn.value)
	return
}
func (cn *Counter) Running() (running uint64) {
	running = atomic.LoadUint64(&cn.running)
	return
}
func (cn *Counter) Max() (max uint64) {
	max = atomic.LoadUint64(&cn.max)
	return
}
