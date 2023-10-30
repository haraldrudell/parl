/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyByValueMap is a mapping whose keys are provided in value order.
package pmaps

// KeyByValueMap is a mapping whose keys are provided in value order
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
// type KeyByValueMap[K comparable, V constraints.Ordered] struct {
// 	// because tree needs less function, must be pointer
// 	*keyByValueMap[K, V] // Get() Length() Range()
// 	 tree                 *btree.BTreeG[K]
// }

// // NewKeyByValueMap returns a mapping whose keys are provided in value order
// func NewKeyByValueMap[K comparable, V constraints.Ordered]() (mpp *KeyByValueMap[K, V]) {
// 	// the tree holds K that is not btree.Ordered
// 	//	- therefore a less function is always required
// 	mp := &keyByValueMap[K, V]{Map: *NewMap[K, V]()}
// 	return &KeyByValueMap[K, V]{
// 		keyByValueMap: mp,
// 		tree:          btree.NewG[K](BtreeDegree, mp.less),
// 	}
// }

// // Put saves or replaces a mapping
// func (m *KeyByValueMap[K, V]) Put(key K, value V) {

// 	// existing mapping
// 	if existing, hasExisting := m.Get(key); hasExisting {

// 		//no-op: key exist with equal rank
// 		if value == existing {
// 			return // exists with equal sort order return: nothing to do
// 		}

// 		// update: key exists but value sorts differently
// 		//	- remove from sorted index
// 		m.tree.Delete(key)
// 	}

// 	// create mapping or update mapped value
// 	m.Map.Put(key, value)
// 	m.tree.ReplaceOrInsert(key) // create in sort order
// }

// // Delete removes mapping using key K.
// //   - if key K is not mapped, the map is unchanged.
// //   - O(log n)
// func (m *KeyByValueMap[K, V]) Delete(key K) {
// 	if _, ok := m.Map.Get(key); !ok {
// 		return
// 	}
// 	m.Map.Delete(key)
// 	m.tree.Delete(key)
// }

// // Clone returns a shallow clone of the map
// func (m *KeyByValueMap[K, V]) Clone() (clone *KeyByValueMap[K, V]) {
// 	return &KeyByValueMap[K, V]{
// 		keyByValueMap: &keyByValueMap[K, V]{
// 			Map: *m.Map.Clone(),
// 		},
// 		tree: m.tree.Clone(),
// 	}
// }

// // Clear empties the map
// //   - clears by re-initializing the map
// //   - when instead ranging and deleting all keys,
// //     the unused size of the map is retained
// func (m *KeyByValueMap[K, V]) Clear() {
// 	m.Map.Clear()
// 	m.tree.Clear(false)
// }

// // List provides the mapped values in order
// //   - O(n)
// func (m *KeyByValueMap[K, V]) List(n ...int) (list []K) {

// 	// empty map case
// 	var length = m.Map.Length()
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

// // pointed-to struct providing less method
// type keyByValueMap[K comparable, V constraints.Ordered] struct {
// 	Map[K, V] // Get() Length() Range()
// }

// // order keys by their corresponding value
// func (m *keyByValueMap[K, V]) less(a, b K) (aBeforeB bool) {
// 	var aV, _ = m.Map.Get(a)
// 	var bV, _ = m.Map.Get(b)
// 	return aV < bV
// }
