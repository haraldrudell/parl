/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"sync"

	"golang.org/x/exp/constraints"
)

// ThreadSafeOrderedMap is a mapping whose values are provided in order. Thread-safe.
type ThreadSafeOrderedMap[K comparable, V constraints.Ordered] struct {
	lock sync.RWMutex
	OrderedMap[K, V]
}

func NewThreadSafeOrderedMap[K comparable, V constraints.Ordered]() (orderedMap *ThreadSafeOrderedMap[K, V]) {
	return &ThreadSafeOrderedMap[K, V]{OrderedMap: *NewOrderedMap[K, V]()}
}

func (mp *ThreadSafeOrderedMap[K, V]) Get(key K) (value V, ok bool) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMap.Get(key)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeOrderedMap[K, V]) Put(key K, value V) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMap.Put(key, value)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeOrderedMap[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	existing, keyExists := mp.OrderedMap.Get(key)
	wasNewKey = !keyExists
	if keyExists && putIf != nil && !putIf(existing) {
		return // putIf false return: this value should not be updated
	}
	mp.OrderedMap.Put(key, value)

	return
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeOrderedMap[K, V]) Delete(key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMap.Delete(key)
}

// Clear empties the map
func (mp *ThreadSafeOrderedMap[K, V]) Clear() {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.OrderedMap.Clear()
}

func (mp *ThreadSafeOrderedMap[K, V]) Length() (length int) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMap.Length()
}

// Clone returns a shallow clone of the map
func (mp *ThreadSafeOrderedMap[K, V]) Clone() (clone *ThreadSafeOrderedMap[K, V]) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	return &ThreadSafeOrderedMap[K, V]{OrderedMap: *mp.OrderedMap.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *ThreadSafeOrderedMap[K, V]) List(n ...int) (list []V) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.OrderedMap.List(n...)
}
