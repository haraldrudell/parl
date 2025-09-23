/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omap1

import (
	"fmt"
	"maps"
)

// OrderedMap is minimal high-performance insertion-ordered map
//   - using OrderedMap as O(1) ordered set:
//   - — create set using [MakeOrderedMapFromKeys] with zero-sized value-type struct{}
//   - — check for element using [OrderedMap.Contains]
//   - — add to set using key-only [OrderedMap.Put]
//   - — iterate set using [OrderedMap.Traverse] and [OrderedMap.TraverseBackwards]
//   - — printable set-element list from [OrderedMap.KeyStrings]
//   - using OrderedMap as O(1) LRU/MRU cache or circular processing:
//   - — use [OrderedMap.GetAndMakeNewest] or
//   - — [OrderedMap.GetAndMakeOldest]
//   - implementation is Go map with ordering by doubly-linked list
//   - — provides the five native Go map functions: Get Put Delete Length Range
//   - — additionally Clear Clone and order-manipulation methods
//   - — forward and backwards Go iterators
//   - — convenient initializers
//   - no dependencies, minimal allocations, simplest api
//   - more lighweight than the b-tree-based maps of
//     github.com/haraldrudell/parl/omaps
//   - much simpler compared to github.com/wk8/go-ordered-map/v2
//   - — less allocations
//   - — easier api and
//   - — less memory consumption
//   - not thread-safe
//   - because OrderedMap, similar to Go map and slice, does not contain non-pointer
//     atomics or locks:
//   - OrderedMap can be passed-by-value, copied and use make
type OrderedMap[K comparable, V any] struct {
	// swissMap map is pointer internally
	//	- map value is pointer to reduce copying on Get
	//	- any map key or map value must be heap allocated
	swissMap map[K]*mappingNode[K, V]
	// head points to the first node of a doubly-linked list
	//	- nil when list is empty
	//	- each node has Prev and Next pointers and key-value fields
	head *mappingNode[K, V]
	// tail points to the last node of a doubly-linked list
	//	- nil when list is empty
	//	- each node has Prev and Next pointers and key-value fields
	tail *mappingNode[K, V]
}

// MakeOrderedMap makes an ordered map of optional pre-allocated size
//   - size: optional pre-allocated size, only >0 honored
//
// Usage:
//
//	var m1 = omap1.MakeOrderedMap[int, string]()
//	var m2 = omap1.MakeOrderedMap[int, string](100)
func MakeOrderedMap[K comparable, V any](size ...int) (m OrderedMap[K, V]) {

	// s is any requested size
	var s int
	if len(size) > 0 {
		s = size[0]
	}
	if s <= 0 {
		m.swissMap = make(map[K]*mappingNode[K, V])
		return
	}
	// s is 1–

	m.swissMap = make(map[K]*mappingNode[K, V], s)

	return
}

// MakeOrderedMapFromKeys creates an ordered map from a set of keys
//   - creates an ordered set with:
//   - — O(1) access
//   - — ordered traversal
//   - values is the zero-value for V
//   - V can be zero-sized type struct{}
//
// Usage:
//
//	var m = omap1.MakeOrderedMapFromKeys[string, struct{}]([]string{
//	  "key1",
//	  "key2",
//	})
func MakeOrderedMapFromKeys[K comparable, V any](keys []K) (m OrderedMap[K, V]) {
	m = MakeOrderedMap[K, V](len(keys))
	for i := range len(keys) {
		m.Put(keys[i])
	}

	return
}

// MakeOrderedMapFromMappings is initializer creating an
// ordered map from a list of mappings
//
// Usage:
//
//	var m = omap1.MakeOrderedMapFromMappings([]omap1.Mappings[string, int]{{
//	  Key: "key1", Value: 1,
//	},{
//	  Key: "key2", Value: 2,
//	}})
func MakeOrderedMapFromMappings[K comparable, V any](mappings []Mapping[K, V]) (m OrderedMap[K, V]) {
	m = MakeOrderedMap[K, V](len(mappings))
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
func (o *OrderedMap[K, V]) Get(key K) (value V, hasValue bool) {
	var mapping *mappingNode[K, V]
	if mapping, hasValue = o.swissMap[key]; hasValue {
		value = mapping.Value
	}

	return
}

// Put creates or replaces a mapping
//   - key: a new or existing key
//   - value: the value to write to the map
//   - — if value missing, the V zero-value is used
//   - old: any replaced value or the zero-value
//   - hadMapping true: the mapping already existed and was updated
//   - hadMapping false: a new mapping was created
func (o *OrderedMap[K, V]) Put(key K, value ...V) (old V, hadMapping bool) {
	var v V
	if len(value) > 0 {
		v = value[0]
	}
	var mapping *mappingNode[K, V]
	if mapping, hadMapping = o.swissMap[key]; hadMapping {
		old = mapping.Value
		mapping.Value = v
		return
	}

	// insert new value into map
	mapping = &mappingNode[K, V]{
		Key:   key,
		Value: v,
	}
	o.swissMap[key] = mapping

	// append value to doubly-linked list after tail
	o.appendNode(mapping)

	return
}

// Delete removes any matching mapping
//   - key: the mapping to delete
//   - old: any value deleted from the map or zero-value
//   - hadMapping true: a mapping was deleted
func (o *OrderedMap[K, V]) Delete(key K) (old V, hadMapping bool) {

	// check if mapping exists
	//	- to update the list, it must be determined if mapping exists
	var mapping *mappingNode[K, V]
	if mapping, hadMapping = o.swissMap[key]; !hadMapping {
		return
	}
	old = mapping.Value

	// remove from map
	//	-After deletion, the map no longer refers to the key or the value,
	// so any pointers in the value become unreachable from the map,
	// allowing garbage collection.
	// https://stackoverflow.com/posts/39395345/revisions
	delete(o.swissMap, key)

	// remove node from list
	o.removeNode(mapping)

	return
}

// Length returns the current number of mappings
func (o *OrderedMap[K, V]) Length() (length int) { return len(o.swissMap) }

// Traverse returns an iterator that traverses mappings from older to newer
//   - key missing or not found: all values are traversed
//   - key present and found: traversal over key and subsequent mappings
func (o *OrderedMap[K, V]) Traverse(key ...K) (iterator func(yield func(key K, value V) (keepGoing bool))) {

	var mapping *mappingNode[K, V]
	if len(key) > 0 {
		mapping = o.swissMap[key[0]]
	}
	if mapping == nil {
		mapping = o.head
	}
	var traverser = newTraverser(mapping)
	iterator = traverser.traverse
	return
}

// Traverse returns an iterator that traverses mappings from newer to older
//   - key missing or not found: all values are traversed
//   - key present and found: traversal over key and preceding mappings
func (o *OrderedMap[K, V]) TraverseBackwards(key ...K) (iterator func(yield func(key K, value V) (keepGoing bool))) {

	var mapping *mappingNode[K, V]
	if len(key) > 0 {
		mapping = o.swissMap[key[0]]
	}
	if mapping == nil {
		mapping = o.tail
	}
	var traverser = newTraverser(mapping, backwards)
	iterator = traverser.traverse
	return
}

// Contains returns true if a mapping for key exists
func (o *OrderedMap[K, V]) Contains(key K) (contains bool) {
	_, contains = o.swissMap[key]
	return
}

// GoMap returns a Go map of current values
func (o *OrderedMap[K, V]) GoMap() (goMap map[K]V) {
	goMap = make(map[K]V, len(o.swissMap))
	for m := o.head; m != nil; m = m.Next {
		goMap[m.Key] = m.Value
	}
	return
}

// Clone returns a clone of the ordered map
//   - data structures are separate but contains the same keys and values
func (o *OrderedMap[K, V]) Clone() (oMap OrderedMap[K, V]) {
	oMap = MakeOrderedMap[K, V](len(o.swissMap))
	for m := o.head; m != nil; m = m.Next {
		oMap.Put(m.Key, m.Value)
	}
	return
}

// Compact re-allocates internal structures to avoid temporary memory leaks
//   - if the map has been large, say 1M elements, Compact releases
//     temporary memory leaks
func (o *OrderedMap[K, V]) Compact() { o.swissMap = maps.Clone(o.swissMap) }

// Keys returns all keys in order
//   - can be used with [OrderedMap.GoMap] to exported ordered mappings
func (o *OrderedMap[K, V]) Keys() (keys []K) {
	keys = make([]K, len(o.swissMap))
	var i int
	for m := o.head; m != nil; m = m.Next {
		keys[i] = m.Key
		i++
	}
	return
}

// KeyStrings returns all keys as %v strings in order
//   - can be used with [strings.Join]
func (o *OrderedMap[K, V]) KeyStrings() (keyStrings []string) {
	keyStrings = make([]string, len(o.swissMap))
	var i int
	for m := o.head; m != nil; m = m.Next {
		keyStrings[i] = fmt.Sprint(m.Key)
		i++
	}

	return
}

// Clear clears the map releasing allocations
//   - if the map has been large, this reduces temporary memory leaks
func (o *OrderedMap[K, V]) Clear() (didClear bool) {
	didClear = len(o.swissMap) > 0
	if !didClear {
		return
	}
	o.swissMap = make(map[K]*mappingNode[K, V])
	o.head = nil
	o.tail = nil

	return
}

// GetAndMakeNewest returns the key mapping and reorders it to be the newest
//   - key: key for a sought mapping
//   - value: present if hasValue true
//   - hasValue true: the mapping did exist
func (o *OrderedMap[K, V]) GetAndMakeNewest(key K) (value V, hasValue bool) {

	var mapping *mappingNode[K, V]
	if mapping, hasValue = o.swissMap[key]; !hasValue {
		return
	}
	value = mapping.Value

	// move mapping to end of list

	// check if mapping already tail
	if mapping.Next == nil {
		return
	}

	// move mapping to end of list
	o.removeNode(mapping)
	o.appendNode(mapping)

	return
}

// GetAndMakeOldest returns the key mapping and reorders it to be the oldest
//   - key: key for a sought mapping
//   - value: present if hasValue true
//   - hasValue true: the mapping did exist
func (o *OrderedMap[K, V]) GetAndMakeOldest(key K) (value V, hasValue bool) {

	var mapping *mappingNode[K, V]
	if mapping, hasValue = o.swissMap[key]; !hasValue {
		return
	}
	value = mapping.Value

	// check if mapping already first item
	if mapping.Prev == nil {
		return
	}

	// move mapping to head of list
	o.removeNode(mapping)
	o.insertNode(mapping)

	return
}

// MoveAfter moves the mapping associated with key to a new position being
// newer than markKey
//   - didMove true: the move took plave
//   - didMove false: the move was not carried out:
//   - — key and markKey are the same
//   - — either key or markKey is not present in the map
func (o *OrderedMap[K, V]) MoveAfter(key, markKey K) (didMove bool) {

	var hasKey bool
	var mapping *mappingNode[K, V]
	var markMapping *mappingNode[K, V]
	if key == markKey {
		return
	} else if mapping, hasKey = o.swissMap[key]; !hasKey {
		return
	} else if markMapping, hasKey = o.swissMap[markKey]; !hasKey {
		return
	}
	didMove = true

	o.removeNode(mapping)
	o.insertNodeAfter(mapping, markMapping)

	return
}

// MoveAfter moves the mapping associated with key to a new position being
// older than markKey
//   - didMove true: the move took plave
//   - didMove false: the move was not carried out:
//   - — key and markKey are the same
//   - — either key or markKey is not present in the map
func (o *OrderedMap[K, V]) MoveBefore(key, markKey K) (didMove bool) {

	var hasKey bool
	var mapping *mappingNode[K, V]
	var markMapping *mappingNode[K, V]
	if key == markKey {
		return
	} else if mapping, hasKey = o.swissMap[key]; !hasKey {
		return
	} else if markMapping, hasKey = o.swissMap[markKey]; !hasKey {
		return
	}
	didMove = true

	o.removeNode(mapping)
	o.insertNodeBefore(mapping, markMapping)

	return
}

// Oldest returns the key and value for the oldest mapping
// or the zero-values otherwise
//   - key: key for oldest mapping if hasValue true. zero-value otherwise
//   - value: key for oldest mapping if hasValue true. zero-value otherwise
//   - hasValue true: the map is not empty
func (o *OrderedMap[K, V]) Oldest() (key K, value V, hasValue bool) {
	if hasValue = len(o.swissMap) > 0; !hasValue {
		return
	}
	var mapping = o.head
	key = mapping.Key
	value = mapping.Value

	return
}

// Newest returns the key and value for the newest mapping
// or the zero-values otherwise
//   - key: key for newest mapping if hasValue true. zero-value otherwise
//   - value: key for newest mapping if hasValue true. zero-value otherwise
//   - hasValue true: the map is not empty
func (o *OrderedMap[K, V]) Newest() (key K, value V, hasValue bool) {
	if hasValue = len(o.swissMap) > 0; !hasValue {
		return
	}
	var mapping = o.tail
	key = mapping.Key
	value = mapping.Value

	return
}

// appendNode appends a node not in the list
// to the end/newest of the doubly-linked list
func (o *OrderedMap[K, V]) appendNode(node *mappingNode[K, V]) {

	// node Next link is nil

	// node Prev link: current tail is before node
	node.Prev = o.tail

	// Next pointer to node:
	// node Next link: node is after tail
	if preceding := node.Prev; preceding != nil {
		// preceding.Next was nil, is now node
		preceding.Next = node
	} else {
		// list was empty: node is also head
		o.head = node
	}

	// Prev pointer to node:	node is new tail
	o.tail = node
}

// insertNode inserts a node not in the list
// to be first/oldest in the doubly-linked list
func (o *OrderedMap[K, V]) insertNode(node *mappingNode[K, V]) {

	// node Next link: current head is after node
	node.Next = o.head

	// node Prev link is nil

	// Next pointer to node: node is new head
	o.head = node

	// Prev pointer to node
	if succeeding := node.Next; succeeding != nil {
		// succeeding.Prev was nil, is now node
		succeeding.Prev = node
	} else {
		// list was empty: node is also tail
		o.tail = node
	}
}

// removeNode removes a node
// from the doubly-linked list
func (o *OrderedMap[K, V]) removeNode(node *mappingNode[K, V]) {

	// unlink means:
	//	- set the Next pointer pointing to node to node.Next
	//	- set the Prev pointer pointing to node to node.Prev

	// Next pointer:
	if preceding := node.Prev; preceding != nil {
		preceding.Next = node.Next
	} else {
		// node is head: update head
		o.head = node.Next
	}

	// Prev pointer:
	if succeeding := node.Next; succeeding != nil {
		succeeding.Prev = node.Prev
		node.Next = nil
	} else {
		// node is tail: update tail
		o.tail = node.Prev
	}

	// drop node pointer
	if node.Prev != nil {
		node.Prev = nil
	}
}

// insertNodeAfter inserts a node not in the list after node precedingNode
func (o *OrderedMap[K, V]) insertNodeAfter(node, precedingNode *mappingNode[K, V]) {

	// node Next link: node is before the node after precedingNode
	node.Next = precedingNode.Next

	// Prev pointer pointing to node
	if nextNode := node.Next; nextNode != nil {
		// node after precedingNode Prev: node is before node after precedingNode
		nextNode.Prev = node
	} else {
		// node is new tail: update tail
		o.tail = node
	}

	// node Prev link: node is after precedingNode
	node.Prev = precedingNode

	// Next pointer pointing to node: node is after precedingNode
	precedingNode.Next = node
}

// insertNodeBefore inserts a node not in the list before succeedingNode
func (o *OrderedMap[K, V]) insertNodeBefore(node, succeedingNode *mappingNode[K, V]) {

	// node Next link: node is before succeedingNode
	node.Next = succeedingNode

	// node Prev: node is after node preceding succeedingNode
	node.Prev = succeedingNode.Prev

	// Next pointer pointing to node: node is after node preceding succeedingNode
	if preceding := succeedingNode.Prev; preceding != nil {
		preceding.Next = node
	} else {
		// node is head: update head
		o.head = node
	}

	// Prev pointer pointing to node: node is before succeedingNode
	succeedingNode.Prev = node
}

// mappingNode is what Go map value points to
//   - doubles as node of a doubly-linked list
//   - contains key for [OrderedMap.Newest] [OrderedMap.Oldest] and
//     cloning [OrderedMap.GoMap] [OrderedMap.Clone] [OrderedMap.Keys]
//   - the first node has Prev nil and is pointed-to by OrderedMap.head
//   - the last node has Next nil and is pointed-to by OrderedMap.tail
type mappingNode[K comparable, V any] struct {
	Prev  *mappingNode[K, V]
	Next  *mappingNode[K, V]
	Key   K
	Value V
}
