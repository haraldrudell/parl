/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"sync"
)

// ThreadSafeInsOrderedMap is a mapping whose values are provided in insertion order. Thread-safe.
//   - implemented with RWMutex controlling InsOrderMap
type ThreadSafeInsOrderedMap[K comparable, V any] struct {
	lock sync.RWMutex
	InsOrderedMap[K, V]
}

func NewThreadSafeInsOrderedMap[K comparable, V any]() (orderedMap *ThreadSafeInsOrderedMap[K, V]) {
	return &ThreadSafeInsOrderedMap[K, V]{InsOrderedMap: *NewInsOrderedMap[K, V]()}
}

func (m *ThreadSafeInsOrderedMap[K, V]) Get(key K) (value V, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.InsOrderedMap.Get(key)
}

// Put saves or replaces a mapping
func (m *ThreadSafeInsOrderedMap[K, V]) Put(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.InsOrderedMap.Put(key, value)
}

// Put saves or replaces a mapping
func (m *ThreadSafeInsOrderedMap[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	existing, keyExists := m.InsOrderedMap.Get(key)
	wasNewKey = !keyExists
	if keyExists && putIf != nil && !putIf(existing) {
		return // putIf false return: this value should not be updated
	}
	m.InsOrderedMap.Put(key, value)

	return
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
func (m *ThreadSafeInsOrderedMap[K, V]) Delete(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.InsOrderedMap.Delete(key)
}

// Clear empties the map
func (m *ThreadSafeInsOrderedMap[K, V]) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.InsOrderedMap.Clear()
}

func (m *ThreadSafeInsOrderedMap[K, V]) Length() (length int) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.InsOrderedMap.Length()
}

// Clone returns a shallow clone of the map
func (m *ThreadSafeInsOrderedMap[K, V]) Clone() (clone *ThreadSafeInsOrderedMap[K, V]) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return &ThreadSafeInsOrderedMap[K, V]{InsOrderedMap: *m.InsOrderedMap.Clone()}
}

// List provides the mapped values in order
func (m *ThreadSafeInsOrderedMap[K, V]) List(n ...int) (list []V) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.InsOrderedMap.List(n...)
}
