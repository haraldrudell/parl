/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"unsafe"
)

// AtomicReference holds a typed reference that is accessed atomically.
type AtomicReference[T any] struct {
	reference *T
}

func MakeAtomicReference[T any]() (reference AtomicReference[T]) {
	return AtomicReference[T]{}
}

func (ref *AtomicReference[T]) Get() (reference *T) {
	return (*T)(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ref.reference)),
	))
}

func (ref *AtomicReference[T]) Put(reference *T) (r0 *T) {
	r0 = (*T)(atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ref.reference)),
		unsafe.Pointer(reference),
	))
	return
}
