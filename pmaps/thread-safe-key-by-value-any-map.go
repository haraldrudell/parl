/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ThreadSafeKeyOrderedByValueMap is a mapping whose keys can be provided in value order. Thread-safe.
package pmaps

import (
	"sync"

	"golang.org/x/exp/constraints"
)

// ThreadSafeKeyByValueAnyMap is a mapping whose keys can be provided in value order. Thread-safe.
type ThreadSafeKeyByValueAnyMap[K comparable, O constraints.Ordered, V any] struct {
	lock sync.RWMutex
	KeyByValueAnyMap[K, O, V]
}

func NewThreadSafeKeyByValueAnyMap[K comparable, O constraints.Ordered, V any](
	valueOrderFunc func(value V) (order O),
) (m *ThreadSafeKeyByValueAnyMap[K, O, V]) {
	return &ThreadSafeKeyByValueAnyMap[K, O, V]{
		KeyByValueAnyMap: *NewKeyByValueAnyMap[K](valueOrderFunc),
	}
}

func (mp *ThreadSafeKeyByValueAnyMap[K, O, V]) Get(key K) (value V, ok bool) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyByValueAnyMap.Get(key)
}

// Put saves or replaces a mapping
func (mp *ThreadSafeKeyByValueAnyMap[K, O, V]) Put(key K, value V) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyByValueAnyMap.Put(key, value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeKeyByValueAnyMap[K, O, V]) Delete(key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyByValueAnyMap.Delete(key)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *ThreadSafeKeyByValueAnyMap[K, O, V]) DeleteFirst() (key K) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	keyList := mp.List(1)
	if len(keyList) == 0 {
		return // no oldest key to delete
	}

	key = keyList[0]
	mp.KeyByValueAnyMap.Delete(key)

	return
}

// Clear empties the map
func (mp *ThreadSafeKeyByValueAnyMap[K, O, V]) Clear() {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.KeyByValueAnyMap.Clear()
}

func (mp *ThreadSafeKeyByValueAnyMap[K, O, V]) Length() (length int) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyByValueAnyMap.Length()
}

// Clone returns a shallow clone of the map
func (mp *ThreadSafeKeyByValueAnyMap[K, O, V]) Clone() (clone *ThreadSafeKeyByValueAnyMap[K, O, V]) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return &ThreadSafeKeyByValueAnyMap[K, O, V]{
		KeyByValueAnyMap: *mp.KeyByValueAnyMap.Clone(),
	}
}

// List provides the mapped values in order
//   - O(n)
func (mp *ThreadSafeKeyByValueAnyMap[K, O, V]) List(n ...int) (list []K) {
	mp.lock.RLock()
	defer mp.lock.RUnlock()

	return mp.KeyByValueAnyMap.List(n...)
}
