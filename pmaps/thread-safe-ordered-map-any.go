/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"sync"

	"golang.org/x/exp/constraints"
)

// ThreadSafeOrderedMapAny is a mapping whose values are provided in order. Thread-safe.
type ThreadSafeOrderedMapAny[K comparable, O constraints.Ordered, V any] struct {
	lock sync.RWMutex
	OrderedMapAny[K, O, V]
}

func NewThreadSafeOrderedMapAny[K comparable, O constraints.Ordered, V any](
	valueOrderFunc func(value V) (order O),
) (orderedMap *ThreadSafeOrderedMapAny[K, O, V]) {
	return &ThreadSafeOrderedMapAny[K, O, V]{
		OrderedMapAny: *NewOrderedMapAny[K](valueOrderFunc)}
}

func (mp *ThreadSafeOrderedMapAny[K, O, V]) Get(key K) (value V, ok bool) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMapAny.Get(key)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeOrderedMapAny[K, O, V]) Put(key K, value V) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMapAny.Put(key, value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeOrderedMapAny[K, O, V]) Delete(key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMapAny.Delete(key)
}

// Clear empties the map
func (mp *ThreadSafeOrderedMapAny[K, O, V]) Clear() {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMapAny.Clear()
}

func (mp *ThreadSafeOrderedMapAny[K, O, V]) Length() (length int) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMapAny.Length()
}

// Clone returns a shallow clone of the map
func (mp *ThreadSafeOrderedMapAny[K, O, V]) Clone() (clone *ThreadSafeOrderedMapAny[K, O, V]) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	return &ThreadSafeOrderedMapAny[K, O, V]{OrderedMapAny: *mp.OrderedMapAny.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *ThreadSafeOrderedMapAny[K, O, V]) List(n ...int) (list []V) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMapAny.List(n...)
}
