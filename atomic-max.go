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
type AtomicMax[T constraints.Integer] struct {
	// threshold is an optional minimum value for a new max
	//	- valid if greater than 0
	threshold uint64
	// whether [AtomicMax.Value] has been invoked
	hasValue atomic.Bool
	// value is current max or 0 if no value is present
	value atomic.Uint64
}

// NewAtomicMax returns a thread-safe max container
//   - T underlying type must be int
//   - negative values are not allowed and cause panic
func NewAtomicMax[T constraints.Integer](threshold T) (atomicMax *AtomicMax[T]) {
	m := AtomicMax[T]{}
	if threshold != 0 {
		m.threshold = m.tToUint64(threshold)
	}
	return &m
}

// Value updates the container possibly with a new Max value
//   - value cannot be negative
//   - Thread-safe
func (m *AtomicMax[T]) Value(value T) (isNewMax bool) {

	// check value against threshold
	var valueU64 = m.tToUint64(value)
	if valueU64 < m.threshold {
		return // below threshold return: isNewMax false
	}

	// 0 as max case
	if isNewMax = m.hasValue.CompareAndSwap(false, true) && value == 0; isNewMax {
		return // first invocation with 0: isNewMax true
	}

	// check against present value
	var current = m.value.Load()
	if isNewMax = valueU64 > current; !isNewMax {
		return // not a new max return: isNewMax false
	}

	// store the new max
	for {

		// try to write value to *max
		if m.value.CompareAndSwap(current, valueU64) {
			return // new max written return: isNewMax true
		}
		if current = m.value.Load(); current >= valueU64 {
			return // no longer a need to write return: isNewMax true
		}
	}
}

// Max returns current max and a flag whether a value is present
//   - Thread-safe
func (m *AtomicMax[T]) Max() (value T, hasValue bool) {
	value = T(m.value.Load())
	hasValue = m.hasValue.Load()
	return
}

// Max1 returns current maximum whether default or set by Value
//   - Thread-safe
func (m *AtomicMax[T]) Max1() (value T) { return T(m.value.Load()) }

// tToUint64 converts T value to uint64
//   - panic if T value is negative
func (m *AtomicMax[T]) tToUint64(value T) (valueU64 uint64) {
	var err error
	if valueU64, err = ints.Unsigned[uint64](value, ""); err != nil {
		panic(err) // value out of range, ie. negative
	}
	return
}
