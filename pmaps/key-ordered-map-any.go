/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyOrderedMapAny is a mapping of uncomparable keys whose keys are ordered by a key-order function
package pmaps

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// KeyOrderedMapAny is a mapping of uncomparable keys whose keys are ordered by a key-order function
type KeyOrderedMapAny[K any, O constraints.Ordered, V any] struct {
	Map[O, V]
	list         parl.Ordered[K]
	keyOrderFunc func(key K) (order O)
}

func NewKeyOrderedMapAny[K any, O constraints.Ordered, V any](
	keyOrderFunc func(key K) (order O),
) (orderedMap *KeyOrderedMapAny[K, O, V]) {
	if keyOrderFunc == nil {
		panic(perrors.NewPF("keyOrderFunc cannot be nil"))
	}
	m := KeyOrderedMapAny[K, O, V]{Map: *NewMap[O, V](), keyOrderFunc: keyOrderFunc}
	m.list = pslices.NewOrderedAny(m.Cmp)
	return &m
}

// Get returns the value mapped by key or the V zero-value otherwise.
//   - the ok return value is true if a mapping was found.
//   - O(1)
func (mp *KeyOrderedMapAny[K, O, V]) Get(key K) (value V, ok bool) {
	value, ok = mp.Map.Get(mp.keyOrderFunc(key))
	return
}

// Put saves or replaces a mapping
func (mp *KeyOrderedMapAny[K, O, V]) Put(key K, value V) {
	mp.Map.Put(mp.keyOrderFunc(key), value)
	mp.list.Insert(key)
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (mp *KeyOrderedMapAny[K, O, V]) Delete(key K) {
	mp.Map.Delete(mp.keyOrderFunc(key))
	mp.list.Delete(key)
}

// Clone returns a shallow clone of the map
func (mp *KeyOrderedMapAny[K, O, V]) Clone() (clone *KeyOrderedMapAny[K, O, V]) {
	return &KeyOrderedMapAny[K, O, V]{
		Map:          *mp.Map.Clone(),
		list:         mp.list.Clone(),
		keyOrderFunc: mp.keyOrderFunc,
	}
}

// List provides the mapped values in order
//   - O(n)
func (mp *KeyOrderedMapAny[K, O, V]) List(n ...int) (list []K) {
	return mp.list.List(n...)
}

func (mp *KeyOrderedMapAny[K, O, V]) Cmp(a, b K) (result int) {
	aOrder := mp.keyOrderFunc(a)
	bOrder := mp.keyOrderFunc(b)
	if aOrder < bOrder {
		return -1
	} else if aOrder > bOrder {
		return 1
	}
	return 0
}
