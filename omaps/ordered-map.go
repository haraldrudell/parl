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
func NewOrderedMap[K comparable, V btree.Ordered](fieldp ...*OrderedMap[K, V]) (orderedMap *OrderedMap[K, V]) {

	// set orderedMap
	if len(fieldp) > 0 {
		orderedMap = fieldp[0]
	}
	if orderedMap == nil {
		orderedMap = &OrderedMap[K, V]{}
	}

	// initialize all fields
	newOrderedMapFunc(&orderedMap.orderedMapFunc)

	return
}

// NewOrderedMapUintptr returns a map for ~uintptr
func NewOrderedMapUintptr[K comparable, V ~uintptr](fieldp ...*OrderedMap[K, V]) (orderedMap *OrderedMap[K, V]) {

	// set orderedMap
	if len(fieldp) > 0 {
		orderedMap = fieldp[0]
	}
	if orderedMap == nil {
		orderedMap = &OrderedMap[K, V]{}
	}

	// initialize all fields
	newOrderedMapFuncUintptr(LessOrdered[V], &orderedMap.orderedMapFunc)

	return
}

// Put creates or replaces a mapping
func (m *OrderedMap[K, V]) Put(key K, value V) {
	m.orderedMapFunc.btreeMap.put(key, value, LessOrdered[V])
}
