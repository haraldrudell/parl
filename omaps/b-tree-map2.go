/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/google/btree"
)

// btreeMap is private version of [BtreeMap] with its methods being public to final consumer
//   - provides a private put helper to multiple implementations
//   - used by OrderedMapFunc
type btreeMap[K comparable, V any] struct{ BTreeMap[K, V] }

// func newBTreeMap2[K comparable, V btree.Ordered]() (m *btreeMap[K, V]) {
// 	return &btreeMap[K, V]{BTreeMap: *NewBTreeMap[K, V]()}
// }

func newBTreeMap2Any[K comparable, V any](fieldp *btreeMap[K, V], less btree.LessFunc[V]) (m *btreeMap[K, V]) {

	// set m
	if m = fieldp; m == nil {
		m = &btreeMap[K, V]{}
	}

	// initialize all fields
	NewBTreeMapAny[K, V](less, &m.BTreeMap)

	return
}

func (m *btreeMap[K, V]) Clone(goMap ...*map[K]V) (clone *btreeMap[K, V]) {
	// only field is BTreeMap
	//	- need clone returning btreeMap

	// goMap case
	if len(goMap) > 0 {
		if gm := goMap[0]; gm != nil {
			m.BTreeMap.Clone(gm)
			return
		}
	}

	// clone case
	//	- only supports any with less function
	clone = &btreeMap[K, V]{}
	var fieldp = &clone.BTreeMap
	var cloneFrom = &m.BTreeMap
	NewBTreeMapClone(fieldp, cloneFrom)

	return
}

func (m *btreeMap[K, V]) cloneToField(clone *btreeMap[K, V]) {

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
