/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package parl

import (
	"sync/atomic"

	"golang.org/x/exp/constraints"
)

// Atomic64 is a generic 64-bit integer with atomic access
//   - generic for named types of any underlying integer or any built-in integer type
//   - generic version of [atomic.Uint64]
//   - Atomic64[int] is atomic int for all platforms
//   - —
//   - go1.21.5 due to alignment using atomic.align64, Atomic64 must be based on [atomic.Uint64]
type Atomic64[T constraints.Integer] struct{ atomic.Uint64 }

// Load atomically loads and returns the value stored in a.
func (a *Atomic64[T]) Load() (value T) { return T(a.Uint64.Load()) }

// Store atomically stores val into a.
func (a *Atomic64[T]) Store(val T) { a.Uint64.Store(uint64(val)) }

// Swap atomically stores new into x and returns the previous value.
func (a *Atomic64[T]) Swap(new T) (old T) { return T(a.Uint64.Swap(uint64(new))) }

// CompareAndSwap executes the compare-and-swap operation for a.
func (a *Atomic64[T]) CompareAndSwap(old, new T) (swapped bool) {
	return a.Uint64.CompareAndSwap(uint64(old), uint64(new))
}

// Add atomically adds delta to a and returns the new value.
func (a *Atomic64[T]) Add(delta T) (new T) { return T(a.Uint64.Add(uint64(delta))) }
