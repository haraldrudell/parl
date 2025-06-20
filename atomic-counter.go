/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"math"
	"sync/atomic"
)

// AtomicCounter is a uint64 thread-safe counter
//   - [atomic.Uint64] exists for simplified increment semantics and
//     non-wrap-around counter
//   - [AtomicCounter.Inc] [AtomicCounter.Dec] delegating to Add
//   - [AtomicCounter.Inc2] [AtomicCounter.Dec2]: preventing wrap-around
//     using CompareAndSwap mechanic
//   - [AtomicCounter.Set] sets particular value
//   - [AtomicCounter.Value] returns current value
type AtomicCounter atomic.Uint64

// Inc increments with wrap-around. Thread-Safe
//   - value is new value
func (a *AtomicCounter) Inc() (value uint64) { return (*atomic.Uint64)(a).Add(1) }

// Inc2 is increment without wrap-around. Thread-Safe
//   - at math.MaxUint64, increments are ineffective
func (a *AtomicCounter) Inc2() (value uint64, didInc bool) {
	for {
		var beforeValue = (*atomic.Uint64)(a).Load()
		if beforeValue == math.MaxUint64 {
			return // at max
		} else if didInc = (*atomic.Uint64)(a).CompareAndSwap(beforeValue, beforeValue+1); didInc {
			return // inc successful return
		}
	}
}

// Dec is decrement with wrap-around. Thread-Safe
func (a *AtomicCounter) Dec() (value uint64) { return (*atomic.Uint64)(a).Add(math.MaxUint64) }

// Dec2 is decrement with no wrap-around. Thread-Safe
//   - at 0, decrements are ineffective
func (a *AtomicCounter) Dec2() (value uint64, didDec bool) {
	for {
		var beforeValue = (*atomic.Uint64)(a).Load()
		if beforeValue == 0 {
			return // no dec return
		} else if didDec = (*atomic.Uint64)(a).CompareAndSwap(beforeValue, beforeValue-1); didDec {
			return // dec successful return
		}
	}
}

// Add is add with wrap-around. Thread-Safe
func (a *AtomicCounter) Add(value uint64) (newValue uint64) { return (*atomic.Uint64)(a).Add(value) }

// Set sets a new aggregate value. Thread-Safe
func (a *AtomicCounter) Set(value uint64) (oldValue uint64) { return (*atomic.Uint64)(a).Swap(value) }

// Value returns current counter-value
func (a *AtomicCounter) Value() (value uint64) { return (*atomic.Uint64)(a).Load() }
