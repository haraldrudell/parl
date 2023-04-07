/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/ints"
	"golang.org/x/exp/constraints"
)

type AtomicMax[T constraints.Integer] struct {
	value    uint64
	hasValue AtomicBool
}

func NewAtomicMax[T constraints.Integer](value T) (atomicMax *AtomicMax[T]) {
	m := AtomicMax[T]{}
	m.value = uint64(value) // set initial threshold
	return &m
}

func (max *AtomicMax[T]) Value(value T) (isNewMax bool) {

	// check if value is a new max
	valueU64, err := ints.Unsigned[uint64](value, "")
	if err != nil {
		panic(err) // value out of range, ie. negative
	}
	maxU64p := (*uint64)(&max.value)
	current := atomic.LoadUint64(maxU64p)
	if isNewMax = valueU64 > current; !isNewMax {
		return // not a new max return
	}
	max.hasValue.Set()

	// store the new max
	for {

		// try to write value to *max
		if atomic.CompareAndSwapUint64(maxU64p, current, valueU64) {
			return // new max written return
		}
		if current = atomic.LoadUint64(maxU64p); current >= valueU64 {
			return // no longer a need to write return
		}
	}
}

func (max *AtomicMax[T]) Max() (value T, hasValue bool) {
	value = T(atomic.LoadUint64((*uint64)(&max.value)))
	hasValue = max.hasValue.IsTrue()
	return
}

func (max *AtomicMax[T]) Max1() (value T) {
	return T(atomic.LoadUint64((*uint64)(&max.value)))
}
