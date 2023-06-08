/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// OrderedMapFunc is a mapping whose values are provided in custom order.
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

// OrderedMapFunc is a mapping whose values are provided in custom order.
type OrderedMapFunc[K comparable, V any] struct {
	Map[K, V]
	list parli.Ordered[V]
}

// NewOrderedMapFunc returns a mapping whose values are provided in custom order.
//   - cmp(a, b) returns:
//   - — a negative number if a should be before b
//   - — 0 if a == b
//   - — a positive number if a should be after b
func NewOrderedMapFunc[K comparable, V any](
	cmp func(a, b V) (result int),
) (orderedMap *OrderedMapFunc[K, V]) {
	return &OrderedMapFunc[K, V]{
		Map:  *NewMap[K, V](),
		list: pslices.NewOrderedAny(cmp),
	}
}

// NewOrderedMapFunc2 returns a mapping whose values are provided in custom order.
func NewOrderedMapFunc2[K comparable, V any](
	list parli.Ordered[V],
) (orderedMap *OrderedMapFunc[K, V]) {
	if list == nil {
		panic(perrors.NewPF("list cannot be nil"))
	} else if list.Length() > 0 {
		list.Clear()
	}
	return &OrderedMapFunc[K, V]{
		Map:  *NewMap[K, V](),
		list: list,
	}
}

// Put saves or replaces a mapping
func (m *OrderedMapFunc[K, V]) Put(key K, value V) {
	m.Map.Put(key, value)
	m.list.Insert(value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (m *OrderedMapFunc[K, V]) Delete(key K) {
	if value, ok := m.Map.Get(key); ok {
		m.Map.Delete(key)
		m.list.Delete(value)
	}
}

// Clone returns a shallow clone of the map
func (m *OrderedMapFunc[K, V]) Clone() (clone *OrderedMapFunc[K, V]) {
	return &OrderedMapFunc[K, V]{Map: *m.Map.Clone(), list: m.list.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (m *OrderedMapFunc[K, V]) List(n ...int) (list []V) {
	return m.list.List(n...)
}
