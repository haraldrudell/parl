/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"sync"

	"golang.org/x/exp/slices"
)

// ThreadSafeSlice provides certain thread-safe operations on a slice
// using [sync.RWMutex] mechanic
//
// thread-safety challenges with a slice:
//   - multiple slice instances may reference the same underlying array
//   - append operation may change address of data and create multiple arrays
//   - data race in reading and writing slice elements
//   - multiple threads accessing a slice instance or their underlying array
//   - due to slicing creating new memory-sharing slices, the slice value must be
//     fully enclosed inside the type providing thread-safety
//   - due to copy and append operations affecting multiple memory addresses,
//     locking is required for thread-safety
//   - maintaining authoritative slice-state during iteration
//
// All slice operations:
//   - len(slice)
//   - append(slice, E...): may create new underlying array
//   - cap(slice)
//   - make([]E, n)
//   - slice[:b]
//   - slice[a:]
//   - slice[a:b]
//   - slice[:]
//   - slice[a]
//   - &slice[a]: address of slice element
//   - &slice: address of slice
//   - copy(slice, slice2)
//   - slice literal
//   - slice variable declaration
//   - assignment of slice element
type ThreadSafeSlice[T any] struct {
	lock  sync.RWMutex
	slice []T
}

func NewThreadSafeSlice[T any]() (threadSafeSlice *ThreadSafeSlice[T]) {
	return &ThreadSafeSlice[T]{}
}

func (t *ThreadSafeSlice[T]) Append(element T) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.slice = append(t.slice, element)
}

func (t *ThreadSafeSlice[T]) Get(index int) (element T, hasValue bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if hasValue = index >= 0 && index < len(t.slice); !hasValue {
		return
	}

	element = t.slice[index]
	return
}

func (t *ThreadSafeSlice[T]) Put(element T, index int) (success bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if success = index >= 0 && index < len(t.slice); !success {
		return
	}

	t.slice[index] = element
	return
}

func (t *ThreadSafeSlice[T]) Length() (length int) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.slice)
}

func (t *ThreadSafeSlice[T]) SliceClone() (clone []T) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return slices.Clone(t.slice)
}

func (t *ThreadSafeSlice[T]) TrimLeft(count int) {
	t.lock.Lock()
	defer t.lock.Unlock()

	TrimLeft(&t.slice, count)
}

func (t *ThreadSafeSlice[T]) SetLength(newLength int) {
	t.lock.Lock()
	defer t.lock.Unlock()

	SetLength(&t.slice, newLength)
}

func (t *ThreadSafeSlice[T]) Clear() {
	t.SetLength(0)
}
