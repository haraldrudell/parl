/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"golang.org/x/exp/maps"
)

type Map[K comparable, V any] struct {
	m map[K]V
}

func NewMap[K comparable, V any]() (mp *Map[K, V]) {
	return &Map[K, V]{m: map[K]V{}}
}

// Get returns the value mapped by key or the V zero-value otherwise.
//   - the ok return value is true if a mapping was found.
//   - O(1)
func (mp *Map[K, V]) Get(key K) (value V, ok bool) {
	value, ok = mp.m[key]
	return
}

// Put saves or replaces a mapping
func (mp *Map[K, V]) Put(key K, value V) {
	mp.m[key] = value
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *Map[K, V]) Delete(key K) {
	delete(mp.m, key)
}

// Clear empties the map
func (mp *Map[K, V]) Clear() {
	maps.Clear(mp.m)
}

// Length returns the number of mappings
func (mp *Map[K, V]) Length() (length int) {
	return len(mp.m)
}

// Clone returns a shallow clone of the map
func (mp *Map[K, V]) Clone() (clone *Map[K, V]) {
	return &Map[K, V]{m: maps.Clone(mp.m)}
}

// List provides the mapped values, undefined ordering
//   - O(n)
func (mp *Map[K, V]) List() (list []V) {
	list = make([]V, len(mp.m))
	i := 0
	for _, v := range mp.m {
		list[i] = v
		i++
	}
	return
}
