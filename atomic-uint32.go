/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import "sync/atomic"

// Uint32 is a 32-bit unsigned integer with atomic access
type Uint32[T ~uint32] struct {
	_ noCopy
	v uint32
}

// Load atomically loads and returns the value stored in x.
func (x *Uint32[T]) Load() T { return T(atomic.LoadUint32(&x.v)) }

// Store atomically stores val into x.
func (x *Uint32[T]) Store(val T) { atomic.StoreUint32(&x.v, uint32(val)) }

// Swap atomically stores new into x and returns the previous value.
func (x *Uint32[T]) Swap(new T) (old T) { return T(atomic.SwapUint32(&x.v, uint32(new))) }

// CompareAndSwap executes the compare-and-swap operation for x.
func (x *Uint32[T]) CompareAndSwap(old, new T) (swapped bool) {
	return atomic.CompareAndSwapUint32(&x.v, uint32(old), uint32(new))
}

// Add atomically adds delta to x and returns the new value.
func (x *Uint32[T]) Add(delta T) (new T) { return T(atomic.AddUint32(&x.v, uint32(delta))) }
