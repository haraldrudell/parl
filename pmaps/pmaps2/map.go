/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pmaps2 contains resusable map types
package pmaps2

import "golang.org/x/exp/maps"

// Map is a reusable promotable Go map
//   - 5 native Go Map functions: Get Put Delete Length Range
//   - convenience functions:
//   - — Clear using fast, scavenging re-create
//   - — Clone using range, optionally appending to provided instance
//   - — these methods require access to the underlying Go map
//   - not thread-safe:
//   - — zero-value delete can be implemented by consumer
//   - — range-Clear with zero-value write can be implemented by consumer
//   - — order methods List and Keys can be implemented by consumer
type Map[K comparable, V any] struct{ goMap map[K]V }

// NewMap returns a reusable Go Map object
func NewMap[K comparable, V any]() (mapping *Map[K, V]) { return &Map[K, V]{goMap: make(map[K]V)} }

// Get returns the value mapped by key or the V zero-value otherwise
//   - ok: true if a mapping was found
//   - O(1)
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	value, ok = m.goMap[key]
	return
}

// Put create or replaces a mapping
func (m *Map[K, V]) Put(key K, value V) { m.goMap[key] = value }

// Delete removes mapping for key
//   - if key is not mapped, the map is unchanged.
//   - O(log n)
func (m *Map[K, V]) Delete(key K) { delete(m.goMap, key) }

// Length returns the number of mappings
func (m *Map[K, V]) Length() (length int) { return len(m.goMap) }

// Range traverses map bindings
//   - iterates over map until rangeFunc returns false
//   - order is undefined
//   - similar to [sync.Map.Range] func (*sync.Map).Range(f func(key any, value any) bool)
func (m *Map[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) (rangedAll bool) {
	for key, value := range m.goMap {
		if !rangeFunc(key, value) {
			return
		}
	}
	return true
}

// Clear empties the map
//   - clears by re-initializing the map
//   - when instead ranging and deleting all keys,
//     the unused size of the map is retained
func (m *Map[K, V]) Clear() { m.goMap = make(map[K]V) }

// Clone returns a shallow clone of the map
//   - mp is an optional pointer to an already allocated map instance
//     to be used and appended to
//   - delegates to [maps.Clone] ranging all keys
func (m *Map[K, V]) Clone(mp ...*Map[K, V]) (clone *Map[K, V]) {

	// clone should point to a destination instance
	if len(mp) > 0 {
		clone = mp[0]
	}
	if clone == nil {
		clone = &Map[K, V]{}
	}

	clone.goMap = maps.Clone(m.goMap)

	return
}
