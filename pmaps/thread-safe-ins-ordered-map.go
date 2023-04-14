/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"sync"
)

// ThreadSafeInsOrderedMap is a mapping whose values are provided in insertion order. Thread-safe.
type ThreadSafeInsOrderedMap[K comparable, V any] struct {
	lock sync.RWMutex
	InsOrderedMap[K, V]
}

func NewThreadSafeInsOrderedMap[K comparable, V any]() (orderedMap *ThreadSafeInsOrderedMap[K, V]) {
	return &ThreadSafeInsOrderedMap[K, V]{InsOrderedMap: *NewInsOrderedMap[K, V]()}
}

func (mp *ThreadSafeInsOrderedMap[K, V]) Get(key K) (value V, ok bool) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.InsOrderedMap.Get(key)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeInsOrderedMap[K, V]) Put(key K, value V) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.InsOrderedMap.Put(key, value)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeInsOrderedMap[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	existing, keyExists := mp.InsOrderedMap.Get(key)
	wasNewKey = !keyExists
	if keyExists && putIf != nil && !putIf(existing) {
		return // putIf false return: this value should not be updated
	}
	mp.InsOrderedMap.Put(key, value)

	return
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeInsOrderedMap[K, V]) Delete(key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.InsOrderedMap.Delete(key)
}

// Clear empties the map
func (mp *ThreadSafeInsOrderedMap[K, V]) Clear() {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.InsOrderedMap.Clear()
}

func (mp *ThreadSafeInsOrderedMap[K, V]) Length() (length int) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.InsOrderedMap.Length()
}

// Clone returns a shallow clone of the map
func (mp *ThreadSafeInsOrderedMap[K, V]) Clone() (clone *ThreadSafeInsOrderedMap[K, V]) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	return &ThreadSafeInsOrderedMap[K, V]{InsOrderedMap: *mp.InsOrderedMap.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *ThreadSafeInsOrderedMap[K, V]) List(n ...int) (list []V) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.InsOrderedMap.List(n...)
}
