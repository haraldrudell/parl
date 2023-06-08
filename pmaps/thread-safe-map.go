/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// RWMap is a one-liner thread-safe mapping.
// RWMap implements [parli.ThreadSafeMap][K comparable, V any].
package pmaps

import (
	"sync"

	"golang.org/x/exp/maps"
)

// ThreadSafeMap is a thread-safe mapping that is reusable
//   - ThreadSafeMap uses reader/writer mutual exclusion lock to attain thread-safety
//   - native functions: Get Put Delete Length Range
//   - convenience functions: Clear Clone (need access to the Go map)
type ThreadSafeMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

// NewThreadSafeMap returns a thread-safe Go map
func NewThreadSafeMap[K comparable, V any]() (pMap *ThreadSafeMap[K, V]) {
	return &ThreadSafeMap[K, V]{m: make(map[K]V)}
}

// Get returns the value mapped by key or the V zero-value otherwise.
//   - the ok return value is true if a mapping was found.
//   - O(1)
func (m *ThreadSafeMap[K, V]) Get(key K) (value V, ok bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	value, ok = m.m[key]

	return
}

// Put saves or replaces a mapping
func (m *RWMap[K, V]) Put(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.m[key] = value
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (m *ThreadSafeMap[K, V]) Delete(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.m, key)
}

// Length returns the number of mappings
func (m *ThreadSafeMap[K, V]) Length() (length int) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.m)
}

// Range traverses map bindings
//   - iterates over map until rangeFunc returns false
//   - similar to: func (*sync.Map).Range(f func(key any, value any) bool)
func (m *ThreadSafeMap[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for k, v := range m.m {
		if !rangeFunc(k, v) {
			return
		}
	}
}

// Clear empties the map
func (m *ThreadSafeMap[K, V]) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.m = make(map[K]V)
}

// Clone returns a shallow clone of the map
func (m *ThreadSafeMap[K, V]) Clone() (clone *ThreadSafeMap[K, V]) {
	var c ThreadSafeMap[K, V]
	clone = &c

	m.lock.RLock()
	defer m.lock.RUnlock()

	c.m = maps.Clone(m.m)

	return
}
