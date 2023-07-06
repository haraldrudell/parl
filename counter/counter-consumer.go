/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

// CounterConsumer is accessible value/running/max values
type CounterConsumer struct {
	// value is the number of Inc invocations and positive Adds
	//	- value will never decrease
	//	- atomic
	value uint64
	//	running is a value going up and down
	//	- the sum of Inc less Dec invocations and positive or negative Adds
	//	- will never go below 0
	//	- atomic
	running uint64 // atomic
	// max is the greatest value running has had
	//	- will never be less than 0
	//	- atomic
	max uint64
	// a controls atomic or lock-based access to value/running/max
	a accessManager
}

var _ parl.CounterValues = &CounterConsumer{} // Counter supports data consumers

// Get returns value/running/max with integrity. Thread-Safe
//   - value is the monotonically increasing value
//   - running is the fluctuating running value
//   - max is the highest value running has had
func (c *CounterConsumer) Get() (value, running, max uint64) {
	defer c.a.Lock().Unlock()

	value = atomic.LoadUint64(&c.value)
	running = atomic.LoadUint64(&c.running)
	max = atomic.LoadUint64(&c.max)

	return
}

// GetReset returns value/running/max with integrity and resets the counter. Thread-Safe
//   - value is the monotonically increasing value
//   - running is the fluctuating running value
//   - max is the highest value running has had
func (c *CounterConsumer) GetReset() (value, running, max uint64) {
	defer c.a.Lock().Unlock()

	value = atomic.SwapUint64(&c.value, 0)
	running = atomic.SwapUint64(&c.running, 0)
	max = atomic.SwapUint64(&c.max, 0)

	return
}

// Value returns the monotonically increasing value
//   - number of Inc invocations and positive Adds
func (c *CounterConsumer) Value() (value uint64) { return atomic.LoadUint64(&c.value) }

// Running returns the fluctuating running value
//   - number of Inc less Dec invocations and sum of Adds
//   - never below 0
func (c *CounterConsumer) Running() (running uint64) { return atomic.LoadUint64(&c.value) }

// Max returns the highest value running has had
func (c *CounterConsumer) Max() (max uint64) { return atomic.LoadUint64(&c.max) }
