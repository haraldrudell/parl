/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyInsOrderedMap is a mapping whose keys are provided in insertion order.
package pmaps

import (
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/google/btree"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/maps"
)

// InsOrderedMap is a mapping whose values are provided in insertion order
//   - updating mapped value for existing key does not update
//     insertion order
//   - native Go Map functions: Get Put Delete Length Range
//   - convenience functions: Clear Clone
//   - order function: List
//   - debug function: Dump
//   - —
//   - ability to return values in insertion order
//   - — can be achieved by retaining keys or values
//   - ability to remove elements from the order by key
//   - B-tree is not optimal because it cannot both maintain order and
//     delete by key
//   - a slice is not optimal because delete leads to large memory copy
//   - therefore:
//   - — as map value, have data value and insertion-order index
//   - — B-tree provides insertion order for those combined values
//   - — delete by key is fast in map and B-tree
type InsOrderedMap[K comparable, V any] struct {
	// store pointer to mapValue so that
	// V can be updated
	m map[K]*mapValue[V]
	// tree provides insertion order
	//	- store pointer to mapValue so it is not duplicated
	tree *btree.BTreeG[*mapValue[V]]
	// insertionIndex counts creation of new mappings
	insertionIndex atomic.Uint64
}

// the map stores pointers to mapValue
type mapValue[V any] struct {
	valuep         atomic.Pointer[V] // thread-safe updatable value
	insertionIndex uint64            // insertion order when mapping was created
}

// NewInsOrderedMap is a mapping whose keys are provided in insertion order.
func NewInsOrderedMap[K comparable, V any]() (orderedMap *InsOrderedMap[K, V]) {
	m := InsOrderedMap[K, V]{m: make(map[K]*mapValue[V])}
	m.tree = btree.NewG(BtreeDegree, m.insOrderLess)
	return &m
}

// Get returns the value mapped by key or the V zero-value otherwise
//   - ok: true if a mapping was found
//   - O(1)
func (m *InsOrderedMap[K, V]) Get(key K) (value V, ok bool) {
	var mapValuep *mapValue[V]
	if mapValuep, ok = m.m[key]; mapValuep != nil {
		if vp := mapValuep.valuep.Load(); vp != nil {
			value = *vp
		}
	}
	return
}

// Put saves or replaces a mapping
func (m *InsOrderedMap[K, V]) Put(key K, value V) {

	// retrieve any existing value
	var mapValuep, ok = m.m[key]

	// new Value
	if !ok {
		// new mapping value with insertion order
		var v = mapValue[V]{insertionIndex: m.insertionIndex.Add(1) - 1}
		v.valuep.Store(&value)
		m.m[key] = &v
		m.tree.ReplaceOrInsert(&v)
		return // new mapping complete return
	}

	// update value
	//	- do not update insertion order
	mapValuep.valuep.Store(&value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
func (m *InsOrderedMap[K, V]) Delete(key K) {

	// retrieve any existing value
	var mapValuep, ok = m.m[key]

	// non-exsiting key case
	if !ok {
		return // key does not exist no-op return
	}

	// delete mapping
	delete(m.m, key)
	m.tree.Delete(mapValuep)
}

// Length returns the number of mappings
func (m *InsOrderedMap[K, V]) Length() (length int) {
	return len(m.m)
}

// Range traverses map bindings
//   - iterates over map until rangeFunc returns false
//   - order is undefined
//   - similar to: func (*sync.Map).Range(f func(key any, value any) bool)
func (m *InsOrderedMap[K, V]) Range(rangeFunc func(key K, value V) (keepGoing bool)) {
	for key, mapValuep := range m.m {
		var value V
		if mapValuep != nil {
			if vp := mapValuep.valuep.Load(); vp != nil {
				value = *vp
			}
		}
		if !rangeFunc(key, value) {
			return
		}
	}
}

// Clear empties the map
//   - clears by re-initializing the map
//   - when instead ranging and deleting all keys,
//     the unused size of the map is retained
func (m *InsOrderedMap[K, V]) Clear() {
	m.m = make(map[K]*mapValue[V])
}

// Clone returns a shallow clone of the map
func (m *InsOrderedMap[K, V]) Clone() (clone *InsOrderedMap[K, V]) {
	c := InsOrderedMap[K, V]{
		m:    maps.Clone(m.m),
		tree: m.tree.Clone(),
	}
	clone.insertionIndex.Store(m.insertionIndex.Load())
	return &c
}

// List provides the mapped values in order
//   - if n is missing or zero, the entire map
//   - otherwise, the first n elements capped by length
func (m *InsOrderedMap[K, V]) List(n ...int) (list []V) {

	// determine length of returned slice
	var nToUse int
	if len(n) > 0 {
		nToUse = n[0]
	}
	if length := len(m.m); nToUse == 0 {
		nToUse = length
	} else if nToUse > length {
		nToUse = length
	}

	// create list slice
	var err error
	if list, err = NewBtreeIterator(m.tree, m.convert).Iterate(nToUse); err != nil {
		// error in converter or nil value pointer
		//	- should never happen
		panic(err)
	}

	return // good return
}

// Dump returns a debug string of ordered values
func (m *InsOrderedMap[K, V]) Dump() (s string) {
	var list = m.List()
	s = "list" + strconv.Itoa(len(list)) + ":"
	for i, v := range list {
		s += fmt.Sprintf("key#%d:%#v-", i, v)
	}
	s += "map" + strconv.Itoa(len(m.m)) + ":"
	for k, v := range m.m {
		s += fmt.Sprintf("key:%#v-value:%#v-", k, v)
	}
	s += "END"
	return
}

// convert returns value type V from mapping value type *mapValue[V]
func (m *InsOrderedMap[K, V]) convert(value *mapValue[V]) (result V, err error) {
	if vp := value.valuep.Load(); vp != nil {
		result = *vp
	} else {
		err = perrors.ErrorfPF("value pointer nil insertion index: %d", value.insertionIndex)
	}
	return
}

// type LessFunc[T any] func(a, b T) bool
var _ btree.LessFunc[int]

// insOrderLess provides insertion-order sort
func (m *InsOrderedMap[K, V]) insOrderLess(a, b *mapValue[V]) (aBeforeB bool) {
	return a.insertionIndex < b.insertionIndex
}
