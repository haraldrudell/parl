/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package imap1

import (
	"fmt"
	"reflect"
	"slices"
)

// IndexedMap provides values with O(1) keyed access and O(1) indexed access
//   - —
//   - IndexedMap combines the fast indexed access and ordered iteration of a slice with
//     the fast keyed access of a map
//   - —
//   - ordering:
//   - slice index can be insertion order, natural sort order for keys or values or
//     sorted according to a provided function based on key and value
//   - sort order may be ascending or descending
//   - for insertion-order, an index could be provided to move an existing
//     mapping in the order or insert a new mapping at a specific location in the order
//   - —
//   - why slice?
//   - [github.com/haraldrudell/parl/pmaps/omap1.OrderedMap] provides
//     ordered map access but not indexed access
//   - — omap1 order is provided by a doubly linked-list that cannot provide fast indexed access
//   - [github.com/haraldrudell/parl/omaps/OrderedMap] is another ordered map with
//     ordering provided by n-ary B-tree.
//   - OrderedMap is more effcient for large number of elements, ie. greater than 1,000
//   - a single slice may cause significant copy of elements upon insert and delete
//   - Go map does not provide order, it provides fast O(1) keyed access
//   - ordering must be provided by a sibling structure to the map
//   - for IndexedMap, doubly-linked list has too low-performing indexed access and
//     B-Tree is too complex for the case of small data structures of up to 1,000 elements
//   - therefore, slice is the sibling structure and the simplest case is a single slice
//   - —
//   - storage strategy for keys and values:
//   - values can be stored as slice element, map value or via pointer
//   - slice elements may be keys, values or pointers to values or value-containers
//   - map values may be values, slice indices, pointer to slice elements or
//     pointers to values or a value-containers
//   - slice indices or pointers to slice may change upon append or delete
//   - therefore, values should not be stored in the slice.
//     The slice should be keys or pointers
//   - if the slice contains keys, then ordered indexed access and iteration is possible
//     for both keys and values
//   - if value-containers are used:
//   - — Put updates to an already mapped key
//     can take place without changes to the map
//   - — the value-container must be independently allocated or a realloc forces significant
//     map changes
//   - — new mappings therefore costs allocation
//   - — if slice index is stored in a value container, deletes cause significant updates
//   - if map values are pointer to value:
//   - — map structures may be smaller and more efficient
//   - — costs allocation for value-types
//   - if the map stores values directly:
//   - — values cannot be directly referenced from the slice
//   - if the map stores slice indices or pointers to slice elements:
//   - — significant map changes occur on delete
//   - therefore:
//   - — map stores the value
//   - — slice elements are keys
//   - — the index for a key or value can only be obtained through
//     linear or other search of the slice
//   - iteration is via index held by consumer
//   - if a value with lower index is concurrently deleted, a value will be
//     missed in the iteration
type IndexedMap[K comparable, V any] struct {
	// keyed provides O(1) keyed access to values
	keyed map[K]V
	// index provide O(1) indexed access according to
	// index’s element order
	index       []K
	vComparable bool
}

// TODO 250922
//   - implement Go Map five native functions: Get Put Delete Length Range
//   - implement methods Clone Clear Compact GoMap TraverseKeys
//     TraverseValues reverse-order-traversal
//   - fill out IndexedMap api to that of [github.com/haraldrudell/parl/pmaps/omap1.OrderedMap]
//   - implement thread-safe version of imap1
//   - implement ordering function for different order than insertion order
//   - implement ordering function to store K any
var TODO int

// MakeIndexedMap makes an ordered map of optional pre-allocated size
//   - size: optional pre-allocated size, only >0 honored
//   - —
//   - ordering is provided by slice: performs best for up to 1,000 elements
//   - for larger maps, consider:
//   - — [github.com/haraldrudell/parl/pmaps/omap1.OrderedMap]
//     ordering by doubly-linked list
//   - — [github.com/haraldrudell/parl/omaps/OrderedMap]
//     ordering provided by n-ary B-tree
//
// Usage:
//
//	var m1 = imap1.MakeIndexedMap[int, string]()
//	var m2 = imap1.MakeIndexedMap[int, string](100)
func MakeIndexedMap[K comparable, V any](size ...int) (m IndexedMap[K, V]) {
	var v V
	m = IndexedMap[K, V]{
		vComparable: reflect.TypeOf(v).Comparable(),
	}

	// s is any requested size
	var s int
	if len(size) > 0 {
		s = size[0]
	}

	if s > 0 {
		m.keyed = make(map[K]V, s)
		m.index = make([]K, s)
	} else {
		m.keyed = make(map[K]V)
	}

	return
}

// Get returns the value mapped by key or the V zero-value otherwise
//   - key: key for a sought mapping
//   - value: present if hasValue true
//   - hasValue true: the mapping did exist
func (m *IndexedMap[K, V]) Get(key K) (value V, hasValue bool) {
	value, hasValue = m.keyed[key]
	return
}

// GetByIndex returns the value mapped by index or the V zero-value otherwise
//   - index: index based on insertion-order 0…Length - 1
//   - value: present if hasValue true
//   - hasValue true: a value for the index did exist
func (m *IndexedMap[K, V]) GetByIndex(index int) (value V, hasValue bool) {
	if index < 0 || index >= len(m.index) {
		return
	}
	return m.Get(m.index[index])
}

// Contains returns true if a mapping for key exists
func (m *IndexedMap[K, V]) Contains(key K) (hasValue bool) {
	_, hasValue = m.keyed[key]
	return
}

// Put creates or replaces a mapping
//   - key: a new or existing key
//   - value: optional value to write to the map
//   - — if value missing, the V zero-value is used
//   - old: any replaced value or the zero-value
//   - hadMapping true: the mapping already existed and was updated
//   - hadMapping false: a new mapping was created
func (m *IndexedMap[K, V]) Put(key K, value ...V) (old V, hadMapping bool) {

	// get new value to use
	var v V
	if len(value) > 0 {
		v = value[0]
	}

	// check if mapping exists
	//	- v and old are type any and cannot be compared
	//	- therefore, write every time
	old, hadMapping = m.keyed[key]
	// try reflect comparable
	if hadMapping && m.vComparable {
		var vNew = reflect.ValueOf(v).Interface()
		var vOld = reflect.ValueOf(old).Interface()
		if vNew == vOld {
			// V is comparable and old and new values are the same
			//	- noop
			return
		}
	}

	// create or update the mapping
	m.keyed[key] = v
	if hadMapping {
		return
	}
	// the mapping did not exist

	// update the index
	m.index = append(m.index, key)

	return
}

// Length returns the current number of mappings
func (m *IndexedMap[K, V]) Length() (length int) { return len(m.index) }

// Delete removes any matching mapping
//   - key: the mapping to delete
//   - old: any value deleted from the map or zero-value
//   - hadMapping true: a mapping was deleted
func (m *IndexedMap[K, V]) Delete(key K) (old V, hadMapping bool) {

	// check if mapping exists
	//	- to update the list, it must be determined if mapping exists
	if old, hadMapping = m.keyed[key]; !hadMapping {
		return
	}

	// remove from map
	//	-After deletion, the map no longer refers to the key or the value,
	// so any pointers in the value become unreachable from the map,
	// allowing garbage collection.
	// https://stackoverflow.com/posts/39395345/revisions
	delete(m.keyed, key)

	// remove key from index
	//	- expensive linear search for K
	if i := slices.Index(m.index, key); i != -1 {
		m.index = slices.Delete(m.index, i, i+1)
	}

	return
}

// Keys returns all keys in order
func (m *IndexedMap[K, V]) Keys() (keys []K) { return slices.Clone(m.index) }

// KeyStrings returns all keys as %v strings in order
//   - can be used with [strings.Join]
func (m *IndexedMap[K, V]) KeyStrings() (keyStrings []string) {
	keyStrings = make([]string, len(m.index))
	for i, key := range m.index {
		keyStrings[i] = fmt.Sprint(key)
	}

	return
}
