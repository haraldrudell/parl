/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package imap1

// TODO 250922
//   - implement thread-safe version
//   - implement ordering function for different order than insertion order
//   - implement ordering function to store K any

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/haraldrudell/parl/pmaps/omap1"
)

// IndexedMap provides both O(1) keyed access and O(1) indexed access.
//   - use IndexedMap when indexed access is required while considering that
//     index-finding operations are expensive and that large IndexedMaps may be less efficient
//   - efficient for small maps of up to 1,000 elements
//   - iteration is by:
//   - — the Values iterator for values, keys and indices: allocation-free, ascending or decsding, full or partial
//   - — consumer using integer for-range with Length/GetByIndex/GetKeyByIndex: allocation-free
//   - — the map itself implementing iterators would cost heap-allocation to hold state
//   - — indexed access is what enables external iteration
//   - ordering by single slice
//   - —
//   - ordering:
//   - generally, slice index can be insertion order, natural sort order for keys or values or
//     sorted according to a provided function based on key and value
//   - sort order may be ascending or descending
//   - for insertion-order, an index could be provided to operations moving an existing
//     mapping in the order or inserting a new mapping at a specific location in the order
//   - —
//   - why slice?
//   - [github.com/haraldrudell/parl/pmaps/omap1.OrderedMap] provides
//     an ordered map without indexed access or a slice
//   - — omap1 order is provided by a doubly linked-list that does not provide efficient indexed access
//   - — the linked list costs allocations
//   - [github.com/haraldrudell/parl/omaps/OrderedMap] is a map with
//     ordering by n-ary B-tree
//   - — OrderedMap is more effcient for large maps, ie. greater than 1,000 elements
//   - — The B-Tree structure is too complex for smaller maps
//   - IndexedMap:
//   - — the single slice may cause significant copy of elements for large maps
//     for insert at specific position and delete
//   - — Delete and other index-finding operations requires linear search of the slice and
//     may incur significant element copy
//   - Go map:
//   - — Go map does not provide order, it provides fast O(1) keyed access
//   - — ordering must be provided by a sibling-structure to the map
//   - IndexedMap’s slice is its sibling structure using the simplest case of a single slice
//   - —
//   - storage strategy for keys and values:
//   - values can be stored as slice element, map value or via pointer
//   - slice elements may be keys, values or pointers to values or value-containers
//   - map values may be values, slice indices, pointer to slice elements or
//     pointers to values or a value-containers
//   - slice indices or pointers to slice may change upon append or delete.
//     Therefore, values should not be stored in the slice.
//     The slice should store keys or pointers
//   - if the slice contains keys, then ordered indexed access and iteration is possible
//     for both keys and values
//   - if value-containers are used:
//   - — Put updates to an already mapped key
//     can take place without changes to the map
//   - — the value-container must be independently allocated or a realloc forces significant
//     map changes
//   - — new mappings therefore costs allocation
//   - — if slice index is stored in a value container, delete causes significant updates
//   - if map values are pointer to value:
//   - — map structures may be smaller and more efficient
//   - — costs allocation for value being a value-type like int
//   - if the map stores values directly:
//   - — values cannot be directly referenced from the slice
//   - if the map stores slice indices or pointers to slice elements:
//   - — significant map changes are required on delete
//   - therefore:
//   - — map stores the value
//   - — slice elements store keys
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

// MakeIndexedMap makes an ordered map of optional pre-allocated size
//   - size: optional pre-allocated size, only >0 honored
//   - —
//   - [IndexedMap.Put] creates or replaces mappings
//   - [IndexedMap.Get] returns values using keyed access
//   - [IndexedMap.GetByIndex] returns values using indexed access
//   - [IndexedMap.GetKeyByIndex] returns keys using indexed access
//   - [Values] is key, vaue and index iterator, ascending or descending
//   - — Values is allocation-free iteration
//   - —
//   - ordering is provided by slice: performs best for up to 1,000 elements
//   - IndexedMap can be stack-allocated pending Go’s escape analysis
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

// MakeIndexedMapFromKeys creates an ordered map from a set of keys
//   - creates an ordered set with:
//   - — O(1) access
//   - — ordered traversal
//   - values is the zero-value for V
//   - V can be zero-sized type struct{}
//
// Usage:
//
//	var m = imap1.MakeIndexedMapFromKeys[string, struct{}]([]string{
//	  "key1",
//	  "key2",
//	})
func MakeIndexedMapFromKeys[K comparable, V any](keys []K) (m IndexedMap[K, V]) {
	m = MakeIndexedMap[K, V](len(keys))
	for i := range len(keys) {
		m.Put(keys[i])
	}

	return
}

// MakeIndexedMapFromMappings is initializer creating an
// ordered map from a list of mappings
//
// Usage:
//
//	var m = omap1.MakeIndexedMapFromMappings([]omap1.Mappings[string, int]{{
//	  Key: "key1", Value: 1,
//	},{
//	  Key: "key2", Value: 2,
//	}})
func MakeIndexedMapFromMappings[K comparable, V any](mappings []omap1.Mapping[K, V]) (m IndexedMap[K, V]) {
	m = MakeIndexedMap[K, V](len(mappings))
	for i := range len(mappings) {
		var mp = &mappings[i]
		m.Put(mp.Key, mp.Value)
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
	value, hasValue = m.Get(m.index[index])
	return
}

// Contains returns true if a mapping for key exists
//   - key: key for a sought mapping
func (m *IndexedMap[K, V]) Contains(key K) (hasValue bool) {
	_, hasValue = m.keyed[key]
	return
}

// GetKeyByIndex returns the key corrresponding to index
//   - index: index based on insertion-order 0…Length - 1
//   - key: the key corresponding to index or zero-value if hasValue false
//   - hasValue true: a value for the index did exist
func (m *IndexedMap[K, V]) GetKeyByIndex(index int) (key K, hasValue bool) {
	if index < 0 || index >= len(m.index) {
		return
	}
	key = m.index[index]
	return
}

// GetIndexForKey returns the index and value for key
//   - key: key for a sought mapping
//   - index: index based on insertion-order 0…Length - 1
//   - value: present if hasValue true
//   - hasValue true: a value for the index did exist
//   - —
//   - requires expensive linear search
func (m *IndexedMap[K, V]) GetIndexForKey(key K) (index int, value V, hasValue bool) {
	if value, hasValue = m.keyed[key]; !hasValue {
		return
	}
	index = slices.Index(m.index, key)

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
//   - —
//   - requires expensive linear search
//   - DeleteByIndex is more performant
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

// DeleteByIndex removes any matching mapping
//   - index: index based on insertion-order 0…Length - 1
//   - old: any value deleted from the map or zero-value
//   - hadMapping true: a mapping was deleted
func (m *IndexedMap[K, V]) DeleteByIndex(index int) (old V, hadMapping bool) {
	if index < 0 || index >= len(m.index) {
		return
	}
	return m.Delete(m.index[index])
}

// LastValue returns the value, key and index for the
// for the most recently added, ie. the newest, mapping
// or the zero-values otherwise
//   - value: key for newest mapping if hasValue true. zero-value otherwise
//   - key: key for newest mapping if hasValue true. zero-value otherwise
//   - index: index based on insertion-order: Length - 1 if hasValue true
//   - hasValue true: the map is not empty
func (m *IndexedMap[K, V]) GetLast() (value V, key K, index int, hasValue bool) {
	if len(m.index) == 0 {
		return
	}
	index = len(m.index) - 1
	key = m.index[index]
	value, hasValue = m.keyed[key]
	return
}

// MoveToIndex moves a mapping in the insertion order from its current position
// to before the element currently at toIndex
//   - fromIndex: 0…Length - 1
//   - toIndex: 0…Length
//   - didMove true: move completed successfully
//   - didMove false: fromIndex ot toIndex out of range,
//     fromIndex equals toIndex
func (m *IndexedMap[K, V]) MoveToIndex(fromIndex, toIndex int) (didMove bool) {
	if fromIndex < 0 || fromIndex >= len(m.index) {
		return
	} else if toIndex < 0 || toIndex > len(m.index) {
		return
	} else if toIndex == fromIndex {
		return
	}
	didMove = true
	var key = m.index[fromIndex]
	if toIndex < fromIndex {
		copy(m.index[toIndex+1:], m.index[toIndex:fromIndex])
		m.index[toIndex] = key
	} else {
		// fromIndex < toIndex
		copy(m.index[fromIndex:], m.index[fromIndex+1:toIndex])
		m.index[toIndex-1] = key
	}
	return
}

// GetAndMakeNewest returns the key mapping and reorders it to be the newest
//   - key: key for a sought mapping
//   - value: present if hasValue true
//   - hasValue true: the mapping did exist
//   - —
//   - requires expensive linear search
func (m *IndexedMap[K, V]) GetAndMakeNewest(key K) (value V, hasValue bool) {
	if value, hasValue = m.keyed[key]; !hasValue {
		return
	} else if key == m.index[len(m.index)-1] {
		return
	}
	// expensive linear search
	var index = slices.Index(m.index, key)
	m.MoveToIndex(index, len(m.index))
	return
}

// GetAndMakeOldest returns the key mapping and reorders it to be the oldest
//   - key: key for a sought mapping
//   - value: present if hasValue true
//   - hasValue true: the mapping did exist
//   - —
//   - requires expensive linear search
func (m *IndexedMap[K, V]) GetAndMakeOldest(key K) (value V, hasValue bool) {
	if value, hasValue = m.keyed[key]; !hasValue {
		return
	} else if key == m.index[0] {
		return
	}
	// expensive linear search
	var index = slices.Index(m.index, key)
	m.MoveToIndex(index, 0)
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

// Clone returns a clone of the ordered map
//   - data structures are separate but contains the same keys and values
func (m *IndexedMap[K, V]) Clone() (oMap IndexedMap[K, V]) {
	oMap.keyed = maps.Clone(m.keyed)
	oMap.index = slices.Clone(m.index)
	oMap.vComparable = m.vComparable
	return
}

// Clear clears the map releasing allocations
//   - if the map has been large, this reduces temporary memory leaks
func (m *IndexedMap[K, V]) Clear() (didClear bool) {
	didClear = len(m.index) > 0
	if !didClear {
		return
	}
	m.keyed = make(map[K]V)
	m.index = nil

	return
}

// Compact re-allocates internal structures to avoid temporary memory leaks
//   - if the map has been large, say 1M elements, Compact releases
//     temporary memory leaks
func (m *IndexedMap[K, V]) Compact() {
	m.keyed = maps.Clone(m.keyed)
	if len(m.index) == 0 {
		m.index = nil
		return
	}
	m.index = slices.Clone(m.index)
}

// GoMap returns a Go map of current values
func (m *IndexedMap[K, V]) GoMap() (goMap map[K]V) { return maps.Clone(m.keyed) }
