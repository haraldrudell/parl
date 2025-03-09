/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/google/btree"
	"github.com/haraldrudell/parl/parli"
)

// OrderedMapFunc is a mapping whose values are provided in custom order
//   - mapping implementation is Go map
//   - ordering structure is B-tree
//   - B-tree offers:
//   - — avoiding vector-copy of large sorted slices which is slow and
//   - — avoiding linear traversal of linked-lists which is slow and
//   - — is a more efficient structure than binary tree
type OrderedMapFunc[K comparable, V any] struct {
	// Get() Length() Range() Delete() Clear() List()
	btreeMap[K, V]
	// type LessFunc[T any] func(a, b T) bool.
	//   - less(a, b) implements sort order and returns:
	//   - — true if a sorts before b
	//   - — false if a is of equal rank to b, or a is after b
	//   - — a equals b must not return true
	less btree.LessFunc[V]
}

// OrderedMapFunc is parli.OrderedMapFunc
var _ parli.ValueOrderedMap[int, string] = &OrderedMapFunc[int, string]{}

// NewOrderedMapFunc returns a mapping whose values are provided in custom order.
//   - less(a, b) implements sort order and returns:
//   - — true if a sorts before b
//   - — false if a is of equal rank to b, or a is after b
//   - — a equals b must not return true
//   - btree.Ordered does not include ~uintptr
func NewOrderedMapFunc[K comparable, V any](
	less func(a, b V) (aBeforeB bool),
	fieldp ...*OrderedMapFunc[K, V],
) (orderedMap *OrderedMapFunc[K, V]) {

	// set orderedMap
	if len(fieldp) > 0 {
		orderedMap = fieldp[0]
	}
	if orderedMap == nil {
		orderedMap = &OrderedMapFunc[K, V]{}
	}

	// initialize all fields
	newBTreeMap2Any(&orderedMap.btreeMap, less)
	orderedMap.less = less

	return
}

// Put creates or replaces a mapping
func (m *OrderedMapFunc[K, V]) Put(key K, value V) {
	m.btreeMap.put(key, value, m.less)
}

// Clone returns a shallow clone of the map
//   - clone is done by ranging all keys
func (m *OrderedMapFunc[K, V]) Clone(goMap ...*map[K]V) (clone parli.ValueOrderedMap[K, V]) {
	return m.Clone2(goMap...)
}

// Clone2 returns a shallow clone of the map
//   - clone is done by ranging all keys
func (m *OrderedMapFunc[K, V]) Clone2(goMap ...*map[K]V) (clone *OrderedMapFunc[K, V]) {

	// clone to Go map case
	if len(goMap) > 0 {
		if gm := goMap[0]; gm != nil {
			m.btreeMap.cloneToGoMap(gm)
			return
		}
	}

	// regular clone case
	clone = &OrderedMapFunc[K, V]{
		less: m.less,
	}
	//btreeMap: *m.btreeMap.Clone(),
	m.btreeMap.cloneToField(&clone.btreeMap)

	return
}
