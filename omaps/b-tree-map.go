/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import (
	"github.com/google/btree"
	"github.com/haraldrudell/parl"
	"golang.org/x/exp/constraints"
)

const (
	BtreeDegree = 6 // each level has 2^6 children: 64
)

// BTreeMap is a reusable, promotable custom value-ordered mapping
//   - 4/5 native Go map functions: Get Delete Length Range
//   - — Put is implemented by consumers that can compare V values
//   - convenience functions:
//   - — Clear using fast, scavenging re-create
//   - — Clone using range, optionally appending to provided instance
//   - order functions:
//   - — List
//   - —
//   - ordering structure is B-tree
//   - B-tree offers:
//   - — avoiding vector-copy of large sorted slices which is slow and
//   - — avoiding linear traversal of linked-lists which is slow and
//   - — is a more efficient structure than binary tree
//   - non-thread-safe maps
//   - mapping implementation is Go map
type BTreeMap[K comparable, V any] struct {
	// Get() Length() Range()
	map2[K, V]
	tree *btree.BTreeG[V]
}

// NewBTreeMap returns a mapping whose values are provided in order
//   - btree.Ordered does not include ~uintptr. For other ordered V, use [NewBTreeMapAny]
func NewBTreeMap[K comparable, V btree.Ordered](fieldp ...*BTreeMap[K, V]) (orderedMap *BTreeMap[K, V]) {
	return newBTreeMap(btree.NewOrderedG[V](BtreeDegree), fieldp...)
}

// NewBTreeMapAny returns a mapping whose values are provided in custom order
//   - supports uintptr. If using default ordered V not uintptr, use [NewBTreeMap]
func NewBTreeMapAny[K comparable, V any](less btree.LessFunc[V], fieldp ...*BTreeMap[K, V]) (orderedMap *BTreeMap[K, V]) {
	parl.NilPanic("less", less)
	return newBTreeMap(btree.NewG(BtreeDegree, less), fieldp...)
}

// NewBTreeMapAny returns a mapping whose values are provided in custom order
//   - fieldp is optional
//   - used for clone to field
//   - less is embedded into tree
func NewBTreeMapClone[K comparable, V any](fieldp *BTreeMap[K, V], cloneFrom *BTreeMap[K, V]) (orderedMap *BTreeMap[K, V]) {
	parl.NilPanic("cloneFrom", cloneFrom)

	// set orderedMap
	if fieldp != nil {
		orderedMap = fieldp
	} else {
		orderedMap = &BTreeMap[K, V]{}
	}

	// handle cloneFrom to all fields
	cloneFrom.map2.cloneToField(&orderedMap.map2)
	orderedMap.tree = cloneFrom.tree.Clone()

	return
}

// newBTreeMap creates using existing B-tree
func newBTreeMap[K comparable, V any](tree *btree.BTreeG[V], fieldp ...*BTreeMap[K, V]) (orderedMap *BTreeMap[K, V]) {

	// set orderedMap
	if len(fieldp) > 0 {
		orderedMap = fieldp[0]
	}
	if orderedMap == nil {
		orderedMap = &BTreeMap[K, V]{}
	}

	// initialize all fields
	newMap(&orderedMap.map2)
	orderedMap.tree = tree

	return
}

// constraints.Ordered includes ~uintptr
type _ interface {
	// ~int | ~int8 | ~int16 | ~int32 | ~int64
	constraints.Signed
	// ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
	constraints.Unsigned
	// ~float32 | ~float64
	constraints.Float
	// ~string
	constraints.Ordered
}

// btree.Ordered does not include ~uintptr
type BtreeOrdered interface {
	// ~int | ~int8 | ~int16 | ~int32 | ~int64 |
	// ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
	// ~float32 | ~float64 | ~string
	btree.Ordered
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (m *BTreeMap[K, V]) Delete(key K) {

	// no-op: delete non-existent mapping
	var existing, hasExisting = m.m2.Get(key)
	if !hasExisting {
		return // maping does not exist return
	}

	// set mapped value to zero value
	var zeroValue V
	m.m2.Put(key, zeroValue)

	// delete mapping
	m.m2.Delete(key)
	m.tree.Delete(existing) // delete from sort order
}

// Clear empties the map
//   - clears by re-initializing the map
//   - when instead ranging and deleting all keys,
//     the unused size of the map is retained
func (m *BTreeMap[K, V]) Clear() {
	m.m2.Clear()
	m.tree.Clear(false)
}

// Clone returns a shallow clone of the map
//   - clone is done by ranging all keys
func (m *BTreeMap[K, V]) Clone(goMap ...*map[K]V) (clone *BTreeMap[K, V]) {

	// Go map clone case
	if len(goMap) > 0 {
		if gm := goMap[0]; gm != nil {
			m.map2.cloneToGoMap(gm)
			return
		}
	}

	// regular clone case
	clone = &BTreeMap[K, V]{
		tree: m.tree.Clone(),
	}
	m.map2.cloneToField(&clone.map2)

	return
}

// List provides mapped values in order
//   - n zero or missing means all items
//   - n non-zero means this many items capped by length
func (m *BTreeMap[K, V]) List(n ...int) (list []V) {

	// empty map case
	var length = m.m2.Length()
	if length == 0 {
		return
	}

	// find non-zero list length [1..length] to use
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
	if list, err = NewBtreeIterator[V, V](m.tree).Iterate(nUse); err != nil {
		panic(err)
	}

	return
}
