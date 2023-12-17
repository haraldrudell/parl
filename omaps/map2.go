/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import "github.com/haraldrudell/parl/pmaps/pmaps2"

// map2 is a private promotable field only promoting
// explicit public identifiers Get Length Range
//   - hiding methods: Clear Clone Delete Put
//   - providing private clone method
//   - innermost map-implementation type for non-thread-safe maps in omaps package
//   - type aliasing does not work for generic types
//   - implementation is [pmaps.Map] wrapping Go map
type map2[K comparable, V any] struct {
	// m2 protects public identifiers from being promoted
	m2 pmaps2.Map[K, V]
}

// map2 is a private promotable field without public identifiers
func newMap[K comparable, V any]() (m *map2[K, V]) { return &map2[K, V]{m2: *pmaps2.NewMap[K, V]()} }

// Get returns the value mapped by key or the V zero-value otherwise
//   - ok: true if a mapping was found
//   - O(1)
func (m *map2[K, V]) Get(key K) (value V, ok bool) { return m.m2.Get(key) }

// Length returns the number of mappings
func (m *map2[K, V]) Length() (length int) { return m.m2.Length() }

// Range traverses map bindings
//   - iterates over map until rangeFunc returns false
//   - order is undefined
//   - similar to [sync.Map.Range] func (*sync.Map).Range(f func(key any, value any) bool)
func (m *map2[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) (rangedAll bool) {
	return m.m2.Range(rangeFunc)
}

func (m *map2[K, V]) clone() (clone *map2[K, V]) { return &map2[K, V]{m2: *m.m2.Clone()} }
