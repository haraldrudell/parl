/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

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
//   - all public methods intended to be public to final consumer
type Map[K comparable, V any] struct{ goMap map[K]V }

// NewMap returns a reusable Go Map object
func NewMap[K comparable, V any]() (mapping *Map[K, V]) { return &Map[K, V]{goMap: make(map[K]V)} }

// NewMap2 is field initializer with optional Go map pointer
//   - fieldp: saves one allocation on field initialization
//   - m: saves one allocation on clone: is the Go map to use
//   - goMap saves allocation on clone to field: is pointer to internal Go map
func NewMap2[K comparable, V any](fieldp *Map[K, V], m map[K]V, goMap ...**map[K]V) (mapping *Map[K, V]) {

	// mapping points to result
	if fieldp != nil {
		mapping = fieldp
	} else {
		mapping = &Map[K, V]{}
	}

	// initialize Go map field
	if m != nil {
		mapping.goMap = m
	} else {
		mapping.goMap = make(map[K]V)
	}

	// check for Go map pointer requested
	var gm **map[K]V
	if len(goMap) == 0 {
		return
	} else if gm = goMap[0]; gm == nil {
		return
	}
	*gm = &mapping.goMap

	return
}

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
func (m *Map[K, V]) Clone(goMap ...*map[K]V) (clone *Map[K, V]) {

	// clone to Go map case
	if len(goMap) > 0 {
		if gm := goMap[0]; gm != nil {
			*gm = maps.Clone(m.goMap)
			return
		}
	}

	// regular clone case
	clone = &Map[K, V]{}
	clone.goMap = maps.Clone(m.goMap)

	return
}
