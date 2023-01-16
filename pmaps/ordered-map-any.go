/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// OrderedMapAny is a mapping of uncomparable keys whose keys are ordered by a key-order function
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// OrderedMapAny is a mapping of uncomparable keys whose keys are ordered by a key-order function
type OrderedMapAny[K comparable, O constraints.Ordered, V any] struct {
	Map[K, V]
	list           parli.Ordered[V]
	valueOrderFunc func(value V) (order O)
}

func NewOrderedMapAny[K comparable, O constraints.Ordered, V any](
	valueOrderFunc func(value V) (order O),
) (orderedMap *OrderedMapAny[K, O, V]) {
	if valueOrderFunc == nil {
		panic(perrors.NewPF("valueOrderFunc cannot be nil"))
	}
	m := OrderedMapAny[K, O, V]{Map: *NewMap[K, V](), valueOrderFunc: valueOrderFunc}
	m.list = pslices.NewOrderedAny(m.Cmp)
	return &m
}

// Get returns the value mapped by key or the V zero-value otherwise.
//   - the ok return value is true if a mapping was found.
//   - O(1)
func (mp *OrderedMapAny[K, O, V]) Get(key K) (value V, ok bool) {
	value, ok = mp.Map.Get(key)
	return
}

// Put saves or replaces a mapping
func (mp *OrderedMapAny[K, O, V]) Put(key K, value V) {
	mp.list.Delete(value)
	mp.Map.Put(key, value)
	mp.list.Insert(value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *OrderedMapAny[K, O, V]) Delete(key K) {
	if value, ok := mp.Map.Get(key); ok {
		mp.Map.Delete(key)
		mp.list.Delete(value)
	}
}

// Clone returns a shallow clone of the map
func (mp *OrderedMapAny[K, O, V]) Clone() (clone *OrderedMapAny[K, O, V]) {
	return &OrderedMapAny[K, O, V]{
		Map:            *mp.Map.Clone(),
		list:           mp.list.Clone(),
		valueOrderFunc: mp.valueOrderFunc,
	}
}

// List provides the mapped values in order
//   - O(n)
func (mp *OrderedMapAny[K, O, V]) List(n ...int) (list []V) {
	return mp.list.List(n...)
}

func (mp *OrderedMapAny[K, O, V]) Cmp(a, b V) (result int) {
	aOrder := mp.valueOrderFunc(a)
	bOrder := mp.valueOrderFunc(b)
	if aOrder < bOrder {
		return -1
	} else if aOrder > bOrder {
		return 1
	}
	return 0
}
