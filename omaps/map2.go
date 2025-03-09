/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/haraldrudell/parl/pmaps/pmaps2"
	"golang.org/x/exp/maps"
)

// map2 is a private promotable field only promoting
// explicit public identifiers Get Length Range
//   - non-thread-safe map implementation for omaps package
//   - hiding methods: Clear Clone Delete Put
//   - providing private clone methods
//   - type aliasing does not work for generic types
//   - implementation is [pmaps2.Map] wrapping Go map
type map2[K comparable, V any] struct {
	// m2 identifier protects public identifiers from being promoted
	m2 pmaps2.Map[K, V]
	// goMap is allocation-saving helper for cloning
	//	- because the referenced map value may change, it has to be pointer to map
	goMap *map[K]V
}

// map2 is a private promotable field without public identifiers
func newMap[K comparable, V any](fieldp *map2[K, V]) (m *map2[K, V]) {

	// set m
	if m = fieldp; m == nil {
		m = &map2[K, V]{}
	}

	// initialize all fields
	var noGoMap map[K]V
	pmaps2.NewMap2(&m.m2, noGoMap, &m.goMap)

	return
}

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

// cloneToField is package-private method providing access to encapsulated Clone
// - clone into existing field without unnecessary alocations
func (m *map2[K, V]) cloneToField(clone *map2[K, V]) {
	// create clone of Go map implementation used by m
	var goMap = maps.Clone(*m.goMap)
	// initialize clone m2 nd goMap fields
	pmaps2.NewMap2(&clone.m2, goMap, &clone.goMap)
}

// cloneToGoMap is package-private method providing access to encapsulated Clone
func (m *map2[K, V]) cloneToGoMap(goMap *map[K]V) { m.m2.Clone(goMap) }
