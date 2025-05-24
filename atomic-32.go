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

// examine [constraints.Unsigned] [constraints.Signed]
type _[T constraints.Unsigned | constraints.Signed] int

// Atomic32 is a generic 32-bit integer with atomic access
//   - generic for named types of select signed and unsigned underlying integers
//   - generic for select built-in integer types
//   - includes ~int ~uint
//   - excludes ~uint64 ~uintptr ~int64
//   - for large values or excluded types, use [Atomic64]
//   - generic version of [atomic.Uint32]
//   - when using int or uint underlying type on a 64-bit platform,
//     type-conversion data loss may occur for larger than 32-bit values
type Atomic32[T uint32Types] struct {
	_ noCopy
	// u32 is a 32-bit unsigned atomic-access field
	u32 atomic.Uint32
}

// atomic.StoreUint32 or atomic.Uint32.Store?
//   - typed atomics are 15% faster
//   - BenchmarkLoad1Uint64 std/sync-atomic/atomic-load-store_bench_test.go

// [atomic.StoreUint32] is C function
var _ = atomic.StoreUint32

// [atomic.Uint32.Store] is C function
var _ = (&atomic.Uint32{}).Store

// Load atomically loads and returns the value stored in a.
func (a *Atomic32[T]) Load() (value T) { return T(a.u32.Load()) }

// Store atomically stores val into a.
func (a *Atomic32[T]) Store(val T) { a.u32.Store(uint32(val)) }

// Swap atomically stores new into a and returns the previous value.
func (a *Atomic32[T]) Swap(new T) (old T) { return T(a.u32.Swap(uint32(new))) }

// CompareAndSwap executes the compare-and-swap operation for a.
func (a *Atomic32[T]) CompareAndSwap(old, new T) (swapped bool) {
	return a.u32.CompareAndSwap(uint32(old), uint32(new))
}

// Or atomically performs a bitwise OR operation on x using the bitmask provided as mask and returns the old value.
func (a *Atomic32[T]) Or(val T) (old T) { return T(a.u32.Or(uint32(val))) }

// And atomically performs a bitwise AND operation on x using the bitmask provided as mask and returns the old value.
func (a *Atomic32[T]) And(val T) (old T) { return T(a.u32.And(uint32(val))) }

// Add atomically adds delta to a and returns the new value.
func (a *Atomic32[T]) Add(delta T) (new T) { return T(a.u32.Add(uint32(delta))) }
