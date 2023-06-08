/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// OrderedMap is a mapping whose values are provided in order.
package pmaps

import (
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// OrderedMap is a mapping whose values are provided in order.
type OrderedMap[K comparable, V constraints.Ordered] struct {
	OrderedMapFunc[K, V] // reusable map with values provided in order
}

// NewOrderedMap returns a mapping whose values are provided in order.
func NewOrderedMap[K comparable, V constraints.Ordered]() (orderedMap *OrderedMap[K, V]) {
	return &OrderedMap[K, V]{
		OrderedMapFunc: *NewOrderedMapFunc2[K, V](pslices.NewOrdered[V]()),
	}
}

// Clone returns a shallow clone of the map
func (m *OrderedMap[K, V]) Clone() (clone *OrderedMap[K, V]) {
	return &OrderedMap[K, V]{OrderedMapFunc: *m.OrderedMapFunc.Clone()}
}
