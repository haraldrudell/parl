/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/google/btree"
	"golang.org/x/exp/constraints"
)

// KeyOrderedMap is a mapping whose keys are provided in order
//   - native Go Map functions: Get Put Delete Length Range
//   - convenience methods: Clear Clone
//   - order method: List
//   - — those methods are implemented because they require access
//     to the underlying Go map
//   - mapping implementation is Go Map
//   - ordering structure is B-tree
//   - B-tree offers:
//   - — avoiding vector-copy of large sorted slices which is slow and
//   - — avoiding linear traversal of linked-lists which is slow and
//   - — is a more efficient structure than binary tree
type KeyOrderedMap[K constraints.Ordered, V any] struct {
	map2[K, V] // Get() Length() Range()
	tree       *btree.BTreeG[K]
}

// NewKeyOrderedMap returns a mapping whose keys are provided in order.
func NewKeyOrderedMap[K btree.Ordered, V any]() (orderedMap *KeyOrderedMap[K, V]) {
	return &KeyOrderedMap[K, V]{
		map2: *newMap[K, V](),
		tree: btree.NewOrderedG[K](BtreeDegree),
	}
}

// NewKeyOrderedMap returns a mapping whose keys are provided in order.
func NewKeyOrderedMapOrdered[K constraints.Ordered, V any]() (orderedMap *KeyOrderedMap[K, V]) {
	return &KeyOrderedMap[K, V]{
		map2: *newMap[K, V](),
		tree: btree.NewG[K](BtreeDegree, Less),
	}
}

func (m *KeyOrderedMap[K, V]) Put(key K, value V) {

	// whether the mapping exists
	//	- value is not comparable, so if mapping exists, the only
	//		action is to overwrite existing value
	var _, hasExisting = m.m2.Get(key)

	// create or update mapping
	m.m2.Put(key, value)

	// if mapping exists, key is already in sort order
	if hasExisting {
		return // key already exists in order return
	}

	m.tree.ReplaceOrInsert(key)
}
func (m *KeyOrderedMap[K, V]) Delete(key K) {

	// no-op: delete non-existent mapping
	if _, hasExisting := m.m2.Get(key); !hasExisting {
		return // maping does not exist return
	}

	// set mapped value to zero value
	var zeroValue V
	m.m2.Put(key, zeroValue)

	// delete mapping
	m.m2.Delete(key)
	m.tree.Delete(key) // delete from sort order
}
func (m *KeyOrderedMap[K, V]) Clear() {
	m.m2.Clear()
	m.tree.Clear(false)
}

// Clone returns a shallow clone of the map
func (m *KeyOrderedMap[K, V]) Clone() (clone *KeyOrderedMap[K, V]) {
	return &KeyOrderedMap[K, V]{
		map2: *m.map2.clone(),
		tree: m.tree.Clone(),
	}
}

// List provides mapped values in order
//   - n zero or missing means all items
//   - n non-zero means this many items capped by length
func (m *KeyOrderedMap[K, V]) List(n ...int) (list []K) {

	// empty map case
	var length = m.m2.Length()
	if length == 0 {
		return
	}

	// non-zero list length [1..length] to use
	var nUse int
	// provided n capped by length
	if len(n) > 0 {
		if nUse = n[0]; nUse > length {
			nUse = length
		}
	}
	// default to full length
	if nUse == 0 {
		nUse = length
	}

	var err error
	if list, err = NewBtreeIterator[K, K](m.tree).Iterate(nUse); err != nil {
		panic(err)
	}

	return
}
