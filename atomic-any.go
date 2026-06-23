// © 2026–present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
// All rights reserved

package parl

import "sync/atomic"

// AtomicAny is generic [atomic.Value] allowing
// to store an interface value without causing allocation.
// In most cases, [atomic.Pointer] is better.
//   - for storing pointers to concrete types, use atomic.Pointer.
//   - AtomicAny can store an interface value of pointer-value
//     runtime type with one less heap-allocation
//     compared to atomic.Pointer
//   - AtomicAny cannot write a nil value
type AtomicAny[T any] struct {
	value atomic.Value
}

func (a *AtomicAny[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return a.value.CompareAndSwap(old, new)
}
func (a *AtomicAny[T]) Load() (value T) {
	value, _ = a.value.Load().(T)
	return
}
func (a *AtomicAny[T]) Store(value T) { a.value.Store(value) }
func (a *AtomicAny[T]) Swap(new T) (old T) {
	old, _ = a.value.Swap(new).(T)
	return
}
