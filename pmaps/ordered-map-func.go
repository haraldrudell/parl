/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// OrderedMapFunc is a mapping whose values are provided in custom order.
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/pslices"
)

// OrderedMapFunc is a mapping whose values are provided in custom order.
type OrderedMapFunc[K comparable, V any] struct {
	Map[K, V]
	list parli.Ordered[V]
}

// NewOrderedMap returns a mapping whose values are provided in order.
func NewOrderedMapFunc[K comparable, V any](
	cmp func(a, b V) (result int),
) (orderedMap *OrderedMapFunc[K, V]) {
	return &OrderedMapFunc[K, V]{
		Map:  *NewMap[K, V](),
		list: pslices.NewOrderedAny(cmp),
	}
}

// Put saves or replaces a mapping
func (mp *OrderedMapFunc[K, V]) Put(key K, value V) {
	mp.Map.Put(key, value)
	mp.list.Insert(value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *OrderedMapFunc[K, V]) Delete(key K) {
	if value, ok := mp.Map.Get(key); ok {
		mp.Map.Delete(key)
		mp.list.Delete(value)
	}
}

// Clone returns a shallow clone of the map
func (mp *OrderedMapFunc[K, V]) Clone() (clone *OrderedMapFunc[K, V]) {
	return &OrderedMapFunc[K, V]{Map: *mp.Map.Clone(), list: mp.list.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *OrderedMapFunc[K, V]) List(n ...int) (list []V) {
	return mp.list.List(n...)
}
