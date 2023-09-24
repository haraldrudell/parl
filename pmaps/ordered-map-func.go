/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// OrderedMapFunc is a mapping whose values are provided in custom order.
package pmaps

import (
	"github.com/google/btree"
)

const (
	btreeDegree = 6 // each level has 2^6 children, 32
)

// OrderedMapFunc is a mapping whose values are provided in custom order
//   - cmp must only return 0 for values that are the same identical value
//   - mapping implementation is Go Map
//   - ordering structure is B-tree
//   - B-tree avoids:
//   - — vector-copy of large sorted slices and
//   - — linear traversal of linked-lists and
//   - — is a more efficient structure than binary tree
type OrderedMapFunc[K comparable, V any] struct {
	Map[K, V] // Get() Length() Range()
	tree      *btree.BTreeG[V]
	cmp       func(a, b V) (result int)
}

var _ btree.LessFunc[int]

// NewOrderedMapFunc returns a mapping whose values are provided in custom order.
//   - cmp(a, b) returns:
//   - — a negative number if a should be before b
//   - — 0 if a == b
//   - — a positive number if a should be after b
func NewOrderedMapFunc[K comparable, V any](
	cmp func(a, b V) (result int),
) (orderedMap *OrderedMapFunc[K, V]) {
	var less = NewCmpLess(cmp)
	return &OrderedMapFunc[K, V]{
		Map:  *NewMap[K, V](),
		tree: btree.NewG[V](btreeDegree, less.Less),
		cmp:  cmp,
	}
}

// NewOrderedMapFunc2 returns a mapping whose values are provided in custom order.
// func NewOrderedMapFunc2[K comparable, V any](
// 	list parli.Ordered[V],
// ) (orderedMap *OrderedMapFunc[K, V]) {
// 	if list == nil {
// 		panic(perrors.NewPF("list cannot be nil"))
// 	} else if list.Length() > 0 {
// 		list.Clear()
// 	}
// 	return &OrderedMapFunc[K, V]{
// 		Map:  *NewMap[K, V](),
// 		list: list,
// 	}
// }

// Put saves or replaces a mapping
func (m *OrderedMapFunc[K, V]) Put(key K, value V) {

	// identical case
	var existing, hasExisting = m.Map.Get(key)
	if hasExisting && m.cmp(value, existing) == 0 {
		return // nothing to do
	}

	// update: key exists but value sorts differently
	//	- remove from sorted index
	if hasExisting {
		m.tree.Delete(existing)
	}

	// new or update
	m.Map.Put(key, value)
	m.tree.ReplaceOrInsert(value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (m *OrderedMapFunc[K, V]) Delete(key K) {

	// does not have case
	var existing, hasExisting = m.Map.Get(key)
	if !hasExisting {
		return
	}

	// delete case
	var zeroValue V
	m.Map.Put(key, zeroValue)
	m.Map.Delete(key)
	m.tree.Delete(existing)
}

// Clear empties the map
//   - re-initialize the map is faster
//   - if ranging and deleting keys, the unused size of the map is retained
func (m *OrderedMapFunc[K, V]) Clear() {
	m.Map.Clear()
	m.tree.Clear(false)
}

// Clone returns a shallow clone of the map
func (m *OrderedMapFunc[K, V]) Clone() (clone *OrderedMapFunc[K, V]) {
	return &OrderedMapFunc[K, V]{
		Map:  *m.Map.Clone(),
		tree: m.tree.Clone(),
		cmp:  m.cmp,
	}
}

// List provides the mapped values in order
//   - n zero or missing means all items
//   - n non-zero means this many items capped by length
func (m *OrderedMapFunc[K, V]) List(n ...int) (list []V) {

	// empty map case
	var length = m.Map.Length()
	if length == 0 {
		return
	}

	// non-zero list length [1..length] to use
	var nUse int
	if len(n) > 0 {
		if nUse = n[0]; nUse > length {
			nUse = length
		}
	}
	if nUse == 0 {
		nUse = length
	}

	list = NewBtreeIterator(m.tree).Iterate(nUse)

	return
}
