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
//   - hasValue indicator true if a value was equal to or greater than threshold
//   - optional threshold for minimum accepted max value
//   - generic for any basic or named Integer type
//   - negative values cause panic
//   - can used to track maximum [time.Duration] that should never be negative
//   - if threshold is not used, initialization-free
//   - —
//   - wait-free CompareAndSwap mechanic
type AtomicMax[T constraints.Integer] struct {
	// threshold is an optional minimum value for a new max
	threshold uint64
	// value is current max
	value atomic.Uint64
	// whether [AtomicMax.Value] has been invoked
	// with value equal or greater to threshold
	hasValue atomic.Bool
}

// NewAtomicMax returns a thread-safe max container
//   - T any basic or named type with underlying type integer
//   - negative values not allowed and cause panic
//   - if threshold is not used, AtomicMax is initialization-free
func NewAtomicMax[T constraints.Integer](threshold T) (atomicMax *AtomicMax[T]) {
	m := AtomicMax[T]{}
	if threshold != 0 {
		m.threshold = m.tToUint64(threshold)
	}
	return &m
}

// Value updates the container with a possible max value
//   - value cannot be negative, that is panic
//   - isNewMax is true if:
//   - — value is equal to or greater than threshold and
//   - — value is the first 0 or
//   - — value was observed to be greater than the current max
//   - upon return, Max and Max1 are guaranteed to return valid data
//   - Thread-safe
func (m *AtomicMax[T]) Value(value T) (isNewMax bool) {

	// check value against threshold
	//	- because no negative values, comparison can be in uint64
	var valueU64 = m.tToUint64(value)
	if valueU64 < m.threshold {
		return // below threshold return: isNewMax false
	}

	// 0 as max case
	var hasValue0 = m.hasValue.Load()
	if valueU64 == 0 {
		if !hasValue0 {
			isNewMax = m.hasValue.CompareAndSwap(false, true)
		}
		return // 0 as max: isNewMax true for first 0 writer
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
			if !hasValue0 {
				// may be rarely written multiple times
				// still faster than CompareAndSwap
				m.hasValue.Store(true)
			}
			return // new max written return: isNewMax true
		}
		if current = m.value.Load(); current >= valueU64 {
			return // no longer a need to write return: isNewMax true
		}
	}
}

// Max returns current max and a flag whether a value is present
//   - Thread-safe
//   - values lack integrity but only matters if Max parallel with first Value
func (m *AtomicMax[T]) Max() (value T, hasValue bool) {
	value = T(m.value.Load())
	hasValue = m.hasValue.Load()
	return
}

// Max1 returns current maximum whether zero-value or set by Value
//   - threshold is ignored
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
