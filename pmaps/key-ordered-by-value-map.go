/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/haraldrudell/parl/parli"

	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// OrderedMap is a mapping whose values are provided in order
type KeyOrderedByValueMap[K comparable, V constraints.Ordered] struct {
	Map[K, V]
	list parli.Ordered[K]
}

// NewOrderedMap returns a mapping whose values are provided in order
func NewKeyOrderedByValueMap[K comparable, V constraints.Ordered]() (m *KeyOrderedByValueMap[K, V]) {
	k := KeyOrderedByValueMap[K, V]{Map: *NewMap[K, V]()}
	k.list = pslices.NewOrderedAny(k.cmp)
	return &k
}

// Put saves or replaces a mapping
func (mp *KeyOrderedByValueMap[K, V]) Put(key K, value V) {
	mp.Map.Put(key, value)
	mp.list.Insert(key)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *KeyOrderedByValueMap[K, V]) Delete(key K) {
	if _, ok := mp.Map.Get(key); ok {
		mp.Map.Delete(key)
		mp.list.Delete(key)
	}
}

// Clone returns a shallow clone of the map
func (mp *KeyOrderedByValueMap[K, V]) Clone() (clone *KeyOrderedByValueMap[K, V]) {
	return &KeyOrderedByValueMap[K, V]{Map: *mp.Map.Clone(), list: mp.list.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *KeyOrderedByValueMap[K, V]) List(n ...int) (list []K) {
	return mp.list.List(n...)
}

// order keys by their corresponding value
func (mp *KeyOrderedByValueMap[K, V]) cmp(a, b K) (result int) {
	aV, _ := mp.Map.Get(a)
	bV, _ := mp.Map.Get(b)
	if aV < bV {
		return -1
	} else if aV > bV {
		return 1
	}
	return 0
}
