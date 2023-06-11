/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
)

type ThreadSafeKeyOrderedMap[K constraints.Ordered, V any] struct {
	ThreadSafeMap[K, V]
	list parli.Ordered[K]
}

// NewThreadSafeKeyOrderedMap returns a mapping whose keys are provided in custom order.
func NewThreadSafeKeyOrderedMap[K constraints.Ordered, V any]() (orderedMap *ThreadSafeKeyOrderedMap[K, V]) {
	return &ThreadSafeKeyOrderedMap[K, V]{
		ThreadSafeMap: *NewThreadSafeMap[K, V](),
		list:          pslices.NewOrdered[K](),
	}
}

// Put saves or replaces a mapping
func (m *ThreadSafeKeyOrderedMap[K, V]) Put(key K, value V) {
	m.ThreadSafeMap.lock.Lock()
	defer m.ThreadSafeMap.lock.Unlock()

	mp := m.ThreadSafeMap.m
	length0 := len(mp)
	mp[key] = value
	if length0 < len(mp) {
		m.list.Insert(key)
	}
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (m *ThreadSafeKeyOrderedMap[K, V]) Delete(key K) {
	m.ThreadSafeMap.lock.Lock()
	defer m.ThreadSafeMap.lock.Unlock()

	delete(m.ThreadSafeMap.m, key)
	m.list.Delete(key)
}

// Clone returns a shallow clone of the map
func (m *ThreadSafeKeyOrderedMap[K, V]) Clear() {
	m.ThreadSafeMap.lock.Lock()
	defer m.ThreadSafeMap.lock.Unlock()

	m.ThreadSafeMap.m = make(map[K]V)
	m.list.Clear()
}

// Clone returns a shallow clone of the map
func (m *ThreadSafeKeyOrderedMap[K, V]) Clone() (clone *ThreadSafeKeyOrderedMap[K, V]) {
	clone = NewThreadSafeKeyOrderedMap[K, V]()
	m.ThreadSafeMap.lock.RLock()
	defer m.ThreadSafeMap.lock.RUnlock()

	clone.ThreadSafeMap.m = maps.Clone(m.ThreadSafeMap.m)
	clone.list = m.list.Clone()
	return
}

// List provides keys in order
//   - O(n)
func (m *ThreadSafeKeyOrderedMap[K, V]) List(n ...int) (list []K) {
	m.ThreadSafeMap.lock.RLock()
	defer m.ThreadSafeMap.lock.RUnlock()

	return m.list.List(n...)
}
