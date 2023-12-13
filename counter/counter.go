/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package counter provides simple and rate counters and tracked datapoints
package counter

import (
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

// Counter is a counter without rate information. Thread-safe.
//   - provider methods: Inc Dec Add
//   - consumer methods: Get GetReset Value Running Max
//   - initialization-free
type Counter struct {
	CounterConsumer
}

var _ parl.Counter = &Counter{}       // Counter supports data providers
var _ parl.CounterValues = &Counter{} // Counter supports data consumers

// newCounter returns a counter without rate information.
func newCounter() (counter parl.Counter) { // Counter is parl.Counter
	return &Counter{}
}

// Inc increments the counter. Thread-Safe, method chaining
func (c *Counter) Inc() (counter parl.Counter) {
	counter = c
	// acquire and relinquish lock or atomic, lock-free access
	defer c.a.RelinquishAccess(c.a.RequestAccess())

	// update value, running and max
	atomic.AddUint64(&c.value, 1)
	c.updateMax(atomic.AddUint64(&c.running, 1)) // returns the new value

	return
}

// Dec decrements the counter but not below zero. Thread-Safe, method chaining
func (c *Counter) Dec() (counter parl.Counter) {
	counter = c
	// acquire and relinquish lock or atomic, lock-free access
	defer c.a.RelinquishAccess(c.a.RequestAccess())

	for {

		// check if running can be decremented
		var running = atomic.LoadUint64(&c.running) // current value
		if running == 0 {
			return // do not decrement lower than 0 return
		}

		// attempt to decrement
		var newRunning = running - 1 // 0…
		if atomic.CompareAndSwapUint64(&c.running, running, newRunning) {
			return // decrement succeeded return
		}
	}
}

// Add adds a positive or negative delta. Thread-Safe, method chaining
func (c *Counter) Add(delta int64) (counter parl.Counter) {
	counter = c
	// check for nothing to do
	if delta == 0 {
		return // no change return
	}
	// acquire and relinquish lock or atomic, lock-free access
	defer c.a.RelinquishAccess(c.a.RequestAccess())

	// positive case: update value, running, max
	if delta > 0 {
		var deltaU64 = uint64(delta)
		atomic.AddUint64(&c.value, deltaU64)
		c.updateMax(atomic.AddUint64(&c.running, deltaU64))
		return // popsitive addition complete return
	}

	// delta negative
	var decrementAmount = uint64(-delta)
	for {

		// cap decrement amount to avoid negative result
		var running = atomic.LoadUint64(&c.running)
		if decrementAmount > running {
			decrementAmount = running
		}

		// check for nothing to do
		if decrementAmount == 0 {
			return // cannot decrement return
		}

		// attempt to store new value
		if atomic.CompareAndSwapUint64(&c.running, running, running-decrementAmount) {
			return // decrement successful return
		}
	}
}

// Consumer return the read-only consumer interface for this counter
func (c *Counter) Consumer() (consumer parl.CounterValues) {
	return &c.CounterConsumer
}

// updateMax ensures that max is at least running
func (c *Counter) updateMax(running uint64) {
	for {

		// check whether max has acceptable value
		var max = atomic.LoadUint64(&c.max) // get current max value
		if running <= max {
			return // no update required return
		}

		// CompareAndSwapUint64 updates to running if value still matches max
		if atomic.CompareAndSwapUint64(&c.max, max, running) {
			return // update successful return
		}
	}
}
