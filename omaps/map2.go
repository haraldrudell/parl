/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import "github.com/haraldrudell/parl/pmaps"

// map2 is a private promotable field only promoting
// explicit public identifiers Get Length Range
type map2[K comparable, V any] struct {
	// m2 protects public identifiers from being promoted
	m2 pmaps.Map[K, V]
}

// map2 is a private promotable field without public identifiers
func newMap[K comparable, V any]() (m *map2[K, V]) {
	return &map2[K, V]{m2: *pmaps.NewMap[K, V]()}
}

func (m *map2[K, V]) Get(key K) (value V, ok bool) {
	return m.m2.Get(key)
}

func (m *map2[K, V]) Length() (length int) {
	return m.m2.Length()
}

func (m *map2[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) {
	m.m2.Range(rangeFunc)
}

func (m *map2[K, V]) clone() (clone *map2[K, V]) {
	return &map2[K, V]{m2: *m.m2.Clone()}
}
