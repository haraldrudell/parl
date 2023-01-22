/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"sync"

	"golang.org/x/exp/constraints"
)

// ThreadSafeKeyOrderedMap is a mapping whose values are provided in order. Thread-safe.
type ThreadSafeKeyOrderedMap[K constraints.Ordered, V constraints.Ordered] struct {
	lock sync.RWMutex
	KeyOrderedMap[K, V]
}

func NewThreadSafeKeyOrderedMap[K constraints.Ordered, V constraints.Ordered]() (orderedMap *ThreadSafeKeyOrderedMap[K, V]) {
	return &ThreadSafeKeyOrderedMap[K, V]{KeyOrderedMap: *NewKeyOrderedMap[K, V]()}
}

func (mp *ThreadSafeKeyOrderedMap[K, V]) Get(key K) (value V, ok bool) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyOrderedMap.Get(key)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeKeyOrderedMap[K, V]) Put(key K, value V) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyOrderedMap.Put(key, value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeKeyOrderedMap[K, V]) Delete(key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyOrderedMap.Delete(key)
}

// Clear empties the map
func (mp *ThreadSafeKeyOrderedMap[K, V]) Clear() {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyOrderedMap.Clear()
}

func (mp *ThreadSafeKeyOrderedMap[K, V]) Length() (length int) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyOrderedMap.Length()
}

// Clone returns a shallow clone of the map
func (mp *ThreadSafeKeyOrderedMap[K, V]) Clone() (clone *ThreadSafeKeyOrderedMap[K, V]) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	return &ThreadSafeKeyOrderedMap[K, V]{KeyOrderedMap: *mp.KeyOrderedMap.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *ThreadSafeKeyOrderedMap[K, V]) List(n ...int) (list []K) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyOrderedMap.List(n...)
}
