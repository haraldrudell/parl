/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyInsOrderedMap is a mapping whose keys are provided in insertion order.
package pmaps

// KeyInsOrderedMap is a mapping whose keys are provided in
// insertion order.
//   - native Go Map functions: Get Put Delete Length Range
//   - convenience methods: Clear Clone
// //   - order method: List
// type KeyInsOrderedMap[K comparable, V any] struct {
// 	// map2 provides O(1) access from key to values
// 	//	- Get() Length() Range()
// 	map2[K, *keyInsertionIndex[K, V]]
// 	//	B-tree provides iteration over keys in insertion order
// 	//	- ability to delete any item
// 	//	- ability to append items
// 	//	- ability to iterate over item in insertion order
// 	//	- clone, clear
// 	//	- a list offers:
// 	//	- — O(log n) binary search and copy meaning expensive delete
// 	//	- — append enlarges tree by complete reallocation
// 	//	- A B-tree of combined insertion-index and key
// 	tree      *btree.BTreeG[*keyInsertionIndex[K, V]]
// 	nextIndex atomic.Uint64
// }

// type keyInsertionIndex[K, V any] struct {
// 	index uint64
// 	key   K
// 	value V
// }

// func less[K, V any](a, b *keyInsertionIndex[K, V]) (aBeforeB bool) {
// 	return a.index < b.index
// }

// var _ btree.LessFunc[*keyInsertionIndex[int, int]] = less[int, int]

// // NewKeyInsOrderedMap is a mapping whose keys are provided in insertion order.
// func NewKeyInsOrderedMap[K constraints.Ordered, V any]() (orderedMap *KeyInsOrderedMap[K, V]) {
// 	return &KeyInsOrderedMap[K, V]{
// 		map2: *newMap[K, *keyInsertionIndex[K, V]](),
// 		tree: btree.NewG[T*keyInsertionIndex[K, V]](BtreeDegree, less[*keyInsertionIndex[K, V]]),
// 	}
// }

// // Put saves or replaces a mapping
// func (m *KeyInsOrderedMap[K, V]) Put(key K, value V) {
// 	if _, ok := m.m2.Get(key); !ok {
// 		m.tree = append(m.tree, key)
// 	}
// 	m.m2.Put(key, value)
// }

// // Delete removes mapping using key K.
// //   - if key K is not mapped, the map is unchanged.
// //   - O(log n)
// func (m *KeyInsOrderedMap[K, V]) Delete(key K) {
// 	m.m2.Delete(key)
// 	if i := slices.Index(m.tree, key); i != -1 {
// 		m.tree = slices.Delete(m.tree, i, i+1)
// 	}
// }

// // Clone returns a shallow clone of the map
// func (m *KeyInsOrderedMap[K, V]) Clone() (clone *KeyInsOrderedMap[K, V]) {
// 	return &KeyInsOrderedMap[K, V]{
// 		map2: *m.map2.clone(),
// 		tree: m.tree.Clone(),
// 	}
// }

// func (m *KeyInsOrderedMap[K, V]) Clear() {
// 	m.m2.Clear()
// 	m.tree.Clear(false)
// }

// // List provides the mapped values in order
// //   - O(n)
// func (m *KeyInsOrderedMap[K, V]) List(n ...int) (list []K) {

// 	// empty map case
// 	var length = m.m2.Length()
// 	if length == 0 {
// 		return
// 	}

// 	// non-zero list length [1..length] to use
// 	var nUse int
// 	// provided n capped by length
// 	if len(n) > 0 {
// 		if nUse = n[0]; nUse > length {
// 			nUse = length
// 		}
// 	}
// 	// default to full length
// 	if nUse == 0 {
// 		nUse = length
// 	}

// 	list = NewBtreeIterator(m.tree).Iterate(nUse)

// 	return
// }
