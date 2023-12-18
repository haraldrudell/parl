/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"sync/atomic"

	"golang.org/x/exp/constraints"
)

// the integer types supported by Atomic32
//   - does not include ~uint64 ~uintptr ~int64
type uint32Types interface {
	~int | ~int8 | ~int16 | ~int32 |
		~uint | ~uint8 | ~uint16 | ~uint32
}

type _[T constraints.Unsigned | constraints.Signed] int

var _ atomic.Uint32

// Atomic32 is a generic 32-bit integer with atomic access
//   - generic for named types of signed and unsigned underlying integers
//   - includes ~int ~uint excluded ~uint64 ~uintptr ~int64
//   - generic version of [atomic.Uint32]
//   - when using int or uint underlying type on a 64-bit platform,
//     type-conversion data loss may occur for larger than 32-bit values
//   - no performance impact compared to other atomics
type Atomic32[T uint32Types] struct {
	_ noCopy
	v uint32
}

// Load atomically loads and returns the value stored in x.
func (x *Atomic32[T]) Load() (value T) { return T(atomic.LoadUint32(&x.v)) }

// Store atomically stores val into x.
func (x *Atomic32[T]) Store(val T) { atomic.StoreUint32(&x.v, uint32(val)) }

// Swap atomically stores new into x and returns the previous value.
func (x *Atomic32[T]) Swap(new T) (old T) { return T(atomic.SwapUint32(&x.v, uint32(new))) }

// CompareAndSwap executes the compare-and-swap operation for x.
func (x *Atomic32[T]) CompareAndSwap(old, new T) (swapped bool) {
	return atomic.CompareAndSwapUint32(&x.v, uint32(old), uint32(new))
}

// Add atomically adds delta to x and returns the new value.
func (x *Atomic32[T]) Add(delta T) (new T) { return T(atomic.AddUint32(&x.v, uint32(delta))) }
