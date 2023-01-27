/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ThreadSafeKeyOrderedByValueMap is a mapping whose keys can be provided in value order. Thread-safe.
package pmaps

import (
	"sync"

	"golang.org/x/exp/constraints"
)

// ThreadSafeKeyOrderedByValueMap is a mapping whose keys can be provided in value order. Thread-safe.
type ThreadSafeKeyOrderedByValueMap[K comparable, V constraints.Ordered] struct {
	lock sync.RWMutex
	KeyOrderedByValueMap[K, V]
}

func NewThreadSafeKeyOrderedByValueMap[K comparable, V constraints.Ordered]() (m *ThreadSafeKeyOrderedByValueMap[K, V]) {
	return &ThreadSafeKeyOrderedByValueMap[K, V]{
		KeyOrderedByValueMap: *NewKeyOrderedByValueMap[K, V](),
	}
}

func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) Get(key K) (value V, ok bool) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyOrderedByValueMap.Get(key)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) Put(key K, value V) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyOrderedByValueMap.Put(key, value)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	existing, keyExists := mp.m[key]
	wasNewKey = !keyExists
	if keyExists && putIf != nil && !putIf(existing) {
		return // putIf false return: this value should not be updated
	}
	mp.m[key] = value

	return
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) Delete(key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyOrderedByValueMap.Delete(key)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) DeleteFirst() (key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	keyList := mp.List(1)
	if len(keyList) == 0 {
		return // no oldest key to delete
	}

	key = keyList[0]
	mp.KeyOrderedByValueMap.Delete(key)

	return
}

// Clear empties the map
func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) Clear() {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyOrderedByValueMap.Clear()
}

func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) Length() (length int) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyOrderedByValueMap.Length()
}

// Clone returns a shallow clone of the map
func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) Clone() (clone *ThreadSafeKeyOrderedByValueMap[K, V]) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return &ThreadSafeKeyOrderedByValueMap[K, V]{
		KeyOrderedByValueMap: *mp.KeyOrderedByValueMap.Clone(),
	}
}

// List provides the mapped values in order
//   - O(n)
func (mp *ThreadSafeKeyOrderedByValueMap[K, V]) List(n ...int) (list []K) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyOrderedByValueMap.List(n...)
}
