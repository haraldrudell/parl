/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/ints"
	"golang.org/x/exp/constraints"
)

// AtomicMax is a thread-safe max container
type AtomicMax[T constraints.Integer] struct{ value, value0 atomic.Uint64 }

// NewAtomicMax returns a thread-safe max container
//   - T underlying type must be int
//   - negative values are not allowed
//   - to set initial value, use Init
func NewAtomicMax[T constraints.Integer]() (atomicMax *AtomicMax[T]) { return &AtomicMax[T]{} }

// Init performs actions that cannot happen prior to copying AtomicMax
//   - supports functional chaining
//   - Thread-safe
func (m *AtomicMax[T]) Init(value T) (atomicMax *AtomicMax[T]) {
	atomicMax = m
	m.value.Store(uint64(value))
	m.value0.Store(uint64(value)) // set initial threshold
	return
}

// Value updates the container possibly with a new Max value
//   - value cannot be negative
//   - Thread-safe
func (m *AtomicMax[T]) Value(value T) (isNewMax bool) {

	// check if value is a new max
	var valueU64, err = ints.Unsigned[uint64](value, "")
	if err != nil {
		panic(err) // value out of range, ie. negative
	}
	var current = m.value.Load()
	if isNewMax = valueU64 > current; !isNewMax {
		return // not a new max return
	}

	// store the new max
	for {

		// try to write value to *max
		if m.value.CompareAndSwap(current, valueU64) {
			return // new max written return
		}
		if current = m.value.Load(); current >= valueU64 {
			return // no longer a need to write return
		}
	}
}

// Max returns current max and a flag whether a value is present
//   - Thread-safe
func (m *AtomicMax[T]) Max() (value T, hasValue bool) {
	var u64 = m.value.Load()
	value = T(u64)
	hasValue = u64 != m.value0.Load()
	return
}

// Max1 returns current maximum whether default or set by Value
//   - Thread-safe
func (m *AtomicMax[T]) Max1() (value T) { return T(m.value.Load()) }
