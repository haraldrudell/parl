/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/haraldrudell/parl"
)

// ThreadSafeInsOrderedMap is a mapping whose values are provided in insertion order. Thread-safe.
//   - implemented with RWMutex controlling InsOrderMap
type ThreadSafeInsOrderedMap[K comparable, V any] struct {
	lock parl.RWMutex
	InsOrderedMap[K, V]
}

func NewThreadSafeInsOrderedMap[K comparable, V any](fieldp ...*ThreadSafeInsOrderedMap[K, V]) (orderedMap *ThreadSafeInsOrderedMap[K, V]) {

	// set orderedMap
	if len(fieldp) > 0 {
		orderedMap = fieldp[0]
	}
	if orderedMap == nil {
		orderedMap = &ThreadSafeInsOrderedMap[K, V]{}
	}

	// initialize all fields
	NewInsOrderedMap[K, V](&orderedMap.InsOrderedMap)

	return
}

func (m *ThreadSafeInsOrderedMap[K, V]) Get(key K) (value V, ok bool) {
	defer m.lock.RLock().RUnlock()

	return m.InsOrderedMap.Get(key)
}

// Put saves or replaces a mapping
func (m *ThreadSafeInsOrderedMap[K, V]) Put(key K, value V) {
	defer m.lock.Lock().Unlock()

	m.InsOrderedMap.Put(key, value)
}

// Put saves or replaces a mapping
func (m *ThreadSafeInsOrderedMap[K, V]) PutIf(key K, value V, putIf func(value V) (doPut bool)) (wasNewKey bool) {
	defer m.lock.Lock().Unlock()

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
	defer m.lock.Lock().Unlock()

	m.InsOrderedMap.Delete(key)
}

// Clear empties the map
func (m *ThreadSafeInsOrderedMap[K, V]) Clear() {
	defer m.lock.Lock().Unlock()

	m.InsOrderedMap.Clear()
}

func (m *ThreadSafeInsOrderedMap[K, V]) Length() (length int) {
	defer m.lock.RLock().RUnlock()

	return m.InsOrderedMap.Length()
}

// Clone returns a shallow clone of the map
func (m *ThreadSafeInsOrderedMap[K, V]) Clone(goMap ...*map[K]V) (clone *ThreadSafeInsOrderedMap[K, V]) {
	var gm *map[K]V
	if len(goMap) > 0 {
		gm = goMap[0]
	}
	if gm == nil {
		clone = &ThreadSafeInsOrderedMap[K, V]{}
	}
	defer m.lock.Lock().Unlock()

	// Go map case
	if gm != nil {
		m.InsOrderedMap.Clone(gm)
		return
	}

	// clone case
	var fieldp = &clone.InsOrderedMap
	var cloneFrom = &m.InsOrderedMap
	NewInsOrderedMapClone(fieldp, cloneFrom)

	return
}

// List provides the mapped values in order
func (m *ThreadSafeInsOrderedMap[K, V]) List(n ...int) (list []V) {
	defer m.lock.RLock().RUnlock()

	return m.InsOrderedMap.List(n...)
}
