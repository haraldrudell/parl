/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"unsafe"
)

type AtomicReferece[T any] struct {
	reference *T
}

func MakeAtomicReferece[T any]() (reference AtomicReferece[T]) {
	return AtomicReferece[T]{}
}

func (ref *AtomicReferece[T]) Get() (reference *T) {
	return (*T)(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ref.reference)),
	))
}

func (ref *AtomicReferece[T]) Put(reference *T) {
	atomic.StorePointer(
		(*unsafe.Pointer)(unsafe.Pointer(&ref.reference)),
		unsafe.Pointer(reference),
	)
}
