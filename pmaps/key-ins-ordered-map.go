/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyInsOrderedMap is a mapping whose keys are provided in insertion order.
package pmaps

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// KeyInsOrderedMap is a mapping whose keys are provided in insertion order.
type KeyInsOrderedMap[K constraints.Ordered, V any] struct {
	Map[K, V]
	list []K
}

// NewKeyInsOrderedMap is a mapping whose keys are provided in insertion order.
func NewKeyInsOrderedMap[K constraints.Ordered, V any]() (orderedMap *KeyInsOrderedMap[K, V]) {
	return &KeyInsOrderedMap[K, V]{Map: *NewMap[K, V]()}
}

// Put saves or replaces a mapping
func (mp *KeyInsOrderedMap[K, V]) Put(key K, value V) {
	if _, ok := mp.Map.Get(key); !ok {
		mp.list = append(mp.list, key)
	}
	mp.Map.Put(key, value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *KeyInsOrderedMap[K, V]) Delete(key K) {
	mp.Map.Delete(key)
	if i := slices.Index(mp.list, key); i != -1 {
		slices.Delete(mp.list, i, i+1)
	}
}

// Clone returns a shallow clone of the map
func (mp *KeyInsOrderedMap[K, V]) Clone() (clone *KeyInsOrderedMap[K, V]) {
	return &KeyInsOrderedMap[K, V]{Map: *mp.Map.Clone(), list: slices.Clone(mp.list)}
}

// List provides the mapped values in order
//   - O(n)
func (mp *KeyInsOrderedMap[K, V]) List(n ...int) (list []K) {

	// default is entire slice
	if len(n) == 0 {
		list = slices.Clone(mp.list)
		return
	}

	// ensure 0 ≦ n ≦ length
	n0 := n[0]
	if n0 <= 0 {
		return
	} else if length := len(mp.list); n0 >= length {
		list = slices.Clone(mp.list)
		return
	}

	list = slices.Clone(mp.list[:n0])

	return
}
