/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/google/btree"
)

// OrderedMapFunc is a mapping whose values are provided in custom order
//   - less(a, b) implements sort order and returns:
//   - — true if a sorts before b
//   - — false if a is of equal rank to b, or a is after b
//   - — a equals b must not return true
//   - mapping implementation is Go Map
//   - ordering structure is B-tree
//   - B-tree offers:
//   - — avoiding vector-copy of large sorted slices which is slow and
//   - — avoiding linear traversal of linked-lists which is slow and
//   - — is a more efficient structure than binary tree
type OrderedMapFunc[K comparable, V any] struct {
	btreeMap[K, V] // Get() Length() Range() Delete() Clear() List()
	// type LessFunc[T any] func(a, b T) bool.
	//   - less(a, b) implements sort order and returns:
	//   - — true if a sorts before b
	//   - — false if a is of equal rank to b, or a is after b
	//   - — a equals b must not return true
	less btree.LessFunc[V]
}

// NewOrderedMapFunc returns a mapping whose values are provided in custom order.
//   - less(a, b) implements sort order and returns:
//   - — true if a sorts before b
//   - — false if a is of equal rank to b, or a is after b
//   - — a equals b must not return true
//   - btree.Ordered does not include ~uintptr
func NewOrderedMapFunc[K comparable, V any](
	less func(a, b V) (aBeforeB bool),
) (orderedMap *OrderedMapFunc[K, V]) {
	return &OrderedMapFunc[K, V]{
		btreeMap: *newBTreeMap2Any[K, V](less),
		less:     less,
	}
}

// Put creates or replaces a mapping
func (m *OrderedMapFunc[K, V]) Put(key K, value V) {
	m.btreeMap.put(key, value, m.sameFunc)
}

// Clone returns a shallow clone of the map
//   - clone is done by ranging all keys
func (m *OrderedMapFunc[K, V]) Clone() (clone *OrderedMapFunc[K, V]) {
	return &OrderedMapFunc[K, V]{
		btreeMap: *m.btreeMap.Clone(),
		less:     m.less,
	}
}

// sameFunc returns if the two values have equal sort-order rank
func (m *OrderedMapFunc[K, V]) sameFunc(value1, value2 V) (isSameRank bool) {
	return !m.less(value1, value2) && !m.less(value2, value1)
}
