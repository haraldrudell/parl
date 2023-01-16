/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"unsafe"
)

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

func (ref *AtomicReference[T]) Put(reference *T) {
	atomic.StorePointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ref.reference)),
		unsafe.Pointer(reference),
	)
}
