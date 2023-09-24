/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import "golang.org/x/exp/maps"

// Map is a Go Map as a reusable promotable field
//   - native Go Map functions: Get Put Delete Length Range
//   - convenience functions: Clear Clone
//   - — those methods are implemented because they require access
//     to the underlying Go map
type Map[K comparable, V any] struct {
	m map[K]V
}

// NewMap returns a resusable Go Map object
func NewMap[K comparable, V any]() (mapping *Map[K, V]) {
	return &Map[K, V]{m: make(map[K]V)}
}

// Get returns the value mapped by key or the V zero-value otherwise.
//   - ok: true if a mapping was found
//   - O(1)
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	value, ok = m.m[key]
	return
}

// Put saves or replaces a mapping
func (m *Map[K, V]) Put(key K, value V) {
	m.m[key] = value
}

// Delete removes mapping for key
//   - if key is not mapped, the map is unchanged.
//   - O(log n)
func (m *Map[K, V]) Delete(key K) {
	delete(m.m, key)
}

// Length returns the number of mappings
func (m *Map[K, V]) Length() (length int) {
	return len(m.m)
}

// Range traverses map bindings
//   - iterates over map until rangeFunc returns false
//   - order is undefined
//   - similar to: func (*sync.Map).Range(f func(key any, value any) bool)
func (m *Map[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) {
	for key, value := range m.m {
		if !rangeFunc(key, value) {
			return
		}
	}
}

// Clear empties the map
//   - re-initialize the map is faster
//   - if ranging and deleting keys, the unused size of the map is retained
func (m *Map[K, V]) Clear() {
	m.m = make(map[K]V)
}

// Clone returns a shallow clone of the map
func (m *Map[K, V]) Clone(mp ...*Map[K, V]) (clone *Map[K, V]) {
	if len(mp) > 0 {
		clone = mp[0]
	}
	if clone == nil {
		clone = &Map[K, V]{}
	}
	clone.m = maps.Clone(m.m)
	return
}
