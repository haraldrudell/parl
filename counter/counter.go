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
//   - value: monotonically increasing affected by method Inc
//   - running: value of Inc - Dec
//   - max: the highest historical value of running
//   - Counter implements parl.Counter and parl.CounterValues.
//
// Note: because Counter uses atomics and not lock, data integrity is not
// guaranteed or achievable. Data is valid and consistent but may be of
// progressing revision.
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

	// update value, running and max
	atomic.AddUint64(&cn.value, 1)
	cn.updateMax(atomic.AddUint64(&cn.running, 1)) // returns the new value

	return
}

func (cn *Counter) Dec() (counter parl.Counter) {
	counter = cn

	for {

		// check if current can be decremented
		current := atomic.LoadUint64(&cn.running) // current value
		if current == 0 {
			return // do not decrement lower than 0 return
		}

		// attenpt to decrement
		newCurrent := current - 1 // 0…
		if atomic.CompareAndSwapUint64(&cn.running, current, newCurrent) {
			return // decrement succeeded return
		}
	}
}

func (cn *Counter) Add(delta int64) (counter parl.Counter) {
	counter = cn

	// check for nothing to do
	if delta == 0 {
		return // no change return
	}

	// positive case: update value, running, max
	if delta > 0 {
		u64 := uint64(delta)
		atomic.AddUint64(&cn.value, u64)
		cn.updateMax(atomic.AddUint64(&cn.running, u64))
		return // popsitive additiona complete return
	}

	// delta negative
	decrementAmount := uint64(-delta)
	for {

		// check current value
		running := atomic.LoadUint64(&cn.running)
		if running == 0 {
			return // cannot decrement further return
		}

		// cap decrement amount to avoid negative result
		var newValue uint64
		if decrementAmount < running {
			newValue = running - decrementAmount
		}

		// attempt to store new value
		if atomic.CompareAndSwapUint64(&cn.running, running, newValue) {
			return // decrement successful return
		}
	}
}

func (cn *Counter) Clone() (counterValues parl.CounterValues) {

	// allocate new container
	cv := Counter{}
	counterValues = &cv

	// copy by reading atomically in order
	cv.value, cv.running, cv.max = cn.Get()
	return
}

func (cn *Counter) CloneReset(stopRateCounters bool) (counterValues parl.CounterValues) {
	counterValues = cn.Clone()

	// clear atomically in order
	atomic.StoreUint64(&cn.value, 0)
	atomic.StoreUint64(&cn.running, 0)
	atomic.StoreUint64(&cn.max, 0)
	return
}

func (cn *Counter) Get() (value, running, max uint64) {
	value = atomic.LoadUint64(&cn.value)
	running = atomic.LoadUint64(&cn.running)
	max = atomic.LoadUint64(&cn.max)
	return
}

func (cn *Counter) Value() (value uint64) {
	value = atomic.LoadUint64(&cn.value)
	return
}

// updateMax ensures that max is at least running
func (cn *Counter) updateMax(running uint64) {
	for {

		// check whether maxc has acceptable value
		max := atomic.LoadUint64(&cn.max) // get current max value
		if running <= max {
			return // no update required return
		}

		// CompareAndSwapUint64 updates to running if value still matches max
		if atomic.CompareAndSwapUint64(&cn.max, max, running) {
			return // update successful return
		}
	}
}
