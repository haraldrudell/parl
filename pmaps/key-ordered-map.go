/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyOrderedMap is a mapping whose keys are provided in order.
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// KeyOrderedMap is a mapping whose keys are provided in order.
type KeyOrderedMap[K constraints.Ordered, V any] struct {
	Map[K, V]
	list parli.Ordered[K]
}

func NewKeyOrderedMap[K constraints.Ordered, V any]() (orderedMap *KeyOrderedMap[K, V]) {
	return &KeyOrderedMap[K, V]{Map: *NewMap[K, V](), list: pslices.NewOrdered[K]()}
}

func newKeyOrderedMap[K constraints.Ordered, V any](list parli.Ordered[K]) (orderedMap *KeyOrderedMap[K, V]) {
	return &KeyOrderedMap[K, V]{Map: *NewMap[K, V](), list: list}
}

// Put saves or replaces a mapping
func (mp *KeyOrderedMap[K, V]) Put(key K, value V) {
	mp.Map.Put(key, value)
	mp.list.Insert(key)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *KeyOrderedMap[K, V]) Delete(key K) {
	mp.Map.Delete(key)
	mp.list.Delete(key)
}

// Clone returns a shallow clone of the map
func (mp *KeyOrderedMap[K, V]) Clone() (clone *KeyOrderedMap[K, V]) {
	return &KeyOrderedMap[K, V]{Map: *mp.Map.Clone(), list: mp.list.Clone()}
}

// List provides the mapped values in order
//   - O(n)
func (mp *KeyOrderedMap[K, V]) List(n ...int) (list []K) {
	return mp.list.List(n...)
}
