/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyInsOrderedMap is a mapping whose keys are provided in insertion order.
package pmaps

import (
	"golang.org/x/exp/slices"
)

// InsOrderedMap is a mapping whose values are provided in insertion order.
type InsOrderedMap[K comparable, V any] struct {
	Map[K, V]
	list []K
}

// NewInsOrderedMap is a mapping whose keys are provided in insertion order.
func NewInsOrderedMap[K comparable, V any]() (orderedMap *InsOrderedMap[K, V]) {
	return &InsOrderedMap[K, V]{Map: *NewMap[K, V]()}
}

// Put saves or replaces a mapping
func (mp *InsOrderedMap[K, V]) Put(key K, value V) {
	if _, ok := mp.Map.Get(key); !ok {
		mp.list = append(mp.list, key)
	}
	mp.Map.Put(key, value)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *InsOrderedMap[K, V]) Delete(key K) {
	if _, ok := mp.Map.Get(key); ok {
		if i := slices.Index(mp.list, key); i != -1 {
			slices.Delete(mp.list, i, i+1)
		}
	}
	mp.Map.Delete(key)
}

// Clone returns a shallow clone of the map
func (mp *InsOrderedMap[K, V]) Clone() (clone *InsOrderedMap[K, V]) {
	return &InsOrderedMap[K, V]{Map: *mp.Map.Clone(), list: slices.Clone(mp.list)}
}

// List provides the mapped values in order
//   - O(n)
func (mp *InsOrderedMap[K, V]) List(n ...int) (list []V) {

	// determine what length to return
	var requestedLength int
	if len(n) == 0 {
		requestedLength = len(mp.list) // default is entire slice
	} else if requestedLength = n[0]; requestedLength < 0 {
		requestedLength = 0 // min 0
	} else if requestedLength > len(mp.list) {
		requestedLength = len(mp.list) // max is len(mp.list)
	}

	list = make([]V, requestedLength)
	for i := 0; i < requestedLength; i++ {
		list[i], _ = mp.Map.Get(mp.list[i])
	}

	return
}
