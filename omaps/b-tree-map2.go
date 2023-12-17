/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/google/btree"
)

// btreeMap is private version of [BtreeMap]
type btreeMap[K comparable, V any] struct{ BTreeMap[K, V] }

// func newBTreeMap2[K comparable, V btree.Ordered]() (m *btreeMap[K, V]) {
// 	return &btreeMap[K, V]{BTreeMap: *NewBTreeMap[K, V]()}
// }

func newBTreeMap2Any[K comparable, V any](less btree.LessFunc[V]) (m *btreeMap[K, V]) {
	return &btreeMap[K, V]{BTreeMap: *NewBTreeMapAny[K, V](less)}
}

func (m *btreeMap[K, V]) Clone() (clone *btreeMap[K, V]) {
	return &btreeMap[K, V]{BTreeMap: *m.BTreeMap.Clone()}
}

func (m *btreeMap[K, V]) put(key K, value V, less btree.LessFunc[V]) {

	// existing mapping
	if existing, hasExisting := m.Get(key); hasExisting {

		//no-op: key exist with equal rank
		//	- if ! value < existing && ! existing < value: values ranked the same
		if !less(value, existing) && !less(existing, value) {
			return // exists with equal sort order return: nothing to do
		}

		// update: key exists but value sorts differently
		//	- remove from sorted index
		m.tree.Delete(existing)
	}

	// create mapping or update mapped value
	m.m2.Put(key, value)
	m.tree.ReplaceOrInsert(value) // create in sort order
}
