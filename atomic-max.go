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

type AtomicMax[T constraints.Integer] struct {
	value    atomic.Uint64
	hasValue atomic.Bool
}

func NewAtomicMax[T constraints.Integer](value T) (atomicMax *AtomicMax[T]) {
	m := AtomicMax[T]{}
	if value != 0 {
		m.value.Store(uint64(value)) // set initial threshold
	}
	return &m
}

func (m *AtomicMax[T]) Value(value T) (isNewMax bool) {

	// check if value is a new max
	valueU64, err := ints.Unsigned[uint64](value, "")
	if err != nil {
		panic(err) // value out of range, ie. negative
	}
	var current = m.value.Load()
	if isNewMax = valueU64 > current; !isNewMax {
		return // not a new max return
	}
	m.hasValue.Store(true)

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

func (m *AtomicMax[T]) Max() (value T, hasValue bool) {
	return T(m.value.Load()), m.hasValue.Load()
}

func (m *AtomicMax[T]) Max1() (value T) {
	return T(m.value.Load())
}
