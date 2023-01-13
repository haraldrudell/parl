/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// OrderedMap is a mapping whose values are provided in order
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// OrderedMap is a mapping whose values are provided in order
type OrderedMap[K comparable, V constraints.Ordered] struct {
	Map[K, V]
	list parli.Ordered[V]
}

// NewOrderedMap returns a mapping whose values are provided in order
func NewOrderedMap[K comparable, V constraints.Ordered]() (orderedMap *OrderedMap[K, V]) {
	return &OrderedMap[K, V]{Map: *NewMap[K, V](), list: pslices.NewOrdered[V]()}
}

// Put saves or replaces a mapping
func (mp *OrderedMap[K, V]) Put(key K, value V) {
	mp.Map.Put(key, value)
	mp.list.Insert(value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *OrderedMap[K, V]) Delete(key K) {
	if value, ok := mp.Map.Get(key); ok {
		mp.Map.Delete(key)
		mp.list.Delete(value)
	}
}

// Clone returns a shallow clone of the map
func (mp *OrderedMap[K, V]) Clone() (clone *OrderedMap[K, V]) {
	return &OrderedMap[K, V]{Map: *mp.Map.Clone(), list: mp.list.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *OrderedMap[K, V]) List(n ...int) (list []V) {
	return mp.list.List(n...)
}
