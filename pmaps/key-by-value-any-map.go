/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/haraldrudell/parl/parli"

	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// KeyByValueAnyMap is a mapping whose values are provided in order
type KeyByValueAnyMap[K comparable, O constraints.Ordered, V any] struct {
	Map[K, V]
	list           parli.Ordered[K]
	valueOrderFunc func(value V) (order O)
}

// NewOrderedMap returns a mapping whose values are provided in order
func NewKeyByValueAnyMap[K comparable, O constraints.Ordered, V any](
	valueOrderFunc func(value V) (order O),
) (m *KeyByValueAnyMap[K, O, V]) {
	k := KeyByValueAnyMap[K, O, V]{Map: *NewMap[K, V]()}
	k.list = pslices.NewOrderedAny(k.cmp)
	return &k
}

// Put saves or replaces a mapping
func (mp *KeyByValueAnyMap[K, O, V]) Put(key K, value V) {
	mp.Map.Put(key, value)
	mp.list.Insert(key)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *KeyByValueAnyMap[K, O, V]) Delete(key K) {
	if _, ok := mp.Map.Get(key); ok {
		mp.Map.Delete(key)
		mp.list.Delete(key)
	}
}

// Clone returns a shallow clone of the map
func (mp *KeyByValueAnyMap[K, O, V]) Clone() (clone *KeyByValueAnyMap[K, O, V]) {
	return &KeyByValueAnyMap[K, O, V]{Map: *mp.Map.Clone(), list: mp.list.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *KeyByValueAnyMap[K, O, V]) List(n ...int) (list []K) {
	return mp.list.List(n...)
}

// order keys by their corresponding value
func (mp *KeyByValueAnyMap[K, O, V]) cmp(a, b K) (result int) {
	aV, _ := mp.Map.Get(a)
	var aOrder O = mp.valueOrderFunc(aV)
	bV, _ := mp.Map.Get(b)
	var bOrder O = mp.valueOrderFunc(bV)
	if aOrder < bOrder {
		return -1
	} else if aOrder > bOrder {
		return 1
	}
	return 0
}
