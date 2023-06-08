/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"golang.org/x/exp/maps"
)

// Map is a Go Map that is reusable
//   - native functions: Get Put Delete Length Range
//   - convenience functions: Clear Clone (need access to the Go map)
type Map[K comparable, V any] struct {
	m map[K]V
}

// NewMap returns a resusable Go Map
func NewMap[K comparable, V any]() (mp *Map[K, V]) {
	return &Map[K, V]{m: make(map[K]V)}
}

// Get returns the value mapped by key or the V zero-value otherwise.
//   - the ok return value is true if a mapping was found.
//   - O(1)
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	value, ok = m.m[key]
	return
}

// Put saves or replaces a mapping
func (m *Map[K, V]) Put(key K, value V) {
	m.m[key] = value
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
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
//   - similar to: func (*sync.Map).Range(f func(key any, value any) bool)
func (m *Map[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) {
	for k, v := range m.m {
		if !rangeFunc(k, v) {
			return
		}
	}
}

// Clear empties the map
func (m *Map[K, V]) Clear() {
	m.m = make(map[K]V)
}

// Clone returns a shallow clone of the map
func (m *Map[K, V]) Clone() (clone *Map[K, V]) {
	return &Map[K, V]{m: maps.Clone(m.m)}
}
