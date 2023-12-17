/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/google/btree"
	"golang.org/x/exp/constraints"
)

// OrderedMap is a mapping whose values are provided in order
//   - mapping implementation is Go Map
//   - ordering structure is B-tree
//   - constraints.Ordered: integer float string
//   - B-tree offers:
//   - — avoiding vector-copy of large sorted slices which is slow and
//   - — avoiding linear traversal of linked-lists which is slow and
//   - — is a more efficient structure than binary tree
type OrderedMap[K comparable, V constraints.Ordered] struct {
	orderedMapFunc[K, V] // Get() Length() Range() Delete() Clear() Clone() List()
}

// NewOrderedMap returns a map for btree.Ordered, ie. not ~uintptr
func NewOrderedMap[K comparable, V btree.Ordered]() (orderedMap *OrderedMap[K, V]) {
	return &OrderedMap[K, V]{orderedMapFunc: *newOrderedMapFunc[K, V]()}
}

// NewOrderedMapUintptr returns a map for ~uintptr
func NewOrderedMapUintptr[K comparable, V ~uintptr]() (orderedMap *OrderedMap[K, V]) {
	return &OrderedMap[K, V]{orderedMapFunc: *newOrderedMapFuncUintptr[K, V](LessOrdered[V])}
}

// Put creates or replaces a mapping
func (m *OrderedMap[K, V]) Put(key K, value V) {
	m.orderedMapFunc.btreeMap.put(key, value, LessOrdered[V])
}
