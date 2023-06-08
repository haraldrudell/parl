/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyOrderedMapFunc is a mapping whose keys are provided in custom order.
package pmaps

import (
	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// KeyOrderedMapFunc is a mapping whose keys are provided in custom order.
type KeyOrderedMapFunc[K comparable, V any] struct {
	Map[K, V]
	list parli.Ordered[K]
}

// NewKeyOrderedMapFunc returns a mapping whose keys are provided in custom order.
func NewKeyOrderedMapFunc[K comparable, V any](
	cmp func(a, b K) (result int),
) (orderedMap *KeyOrderedMapFunc[K, V]) {
	return &KeyOrderedMapFunc[K, V]{
		Map:  *NewMap[K, V](),
		list: pslices.NewOrderedAny(cmp),
	}
}

// NewKeyOrderedMapFunc2 returns a mapping whose keys are provided in order.
func NewKeyOrderedMapFunc2[K constraints.Ordered, V any](
	list parli.Ordered[K],
) (orderedMap *KeyOrderedMapFunc[K, V]) {
	if list == nil {
		panic(perrors.NewPF("list cannot be nil"))
	} else if list.Length() > 0 {
		list.Clear()
	}
	return &KeyOrderedMapFunc[K, V]{
		Map:  *NewMap[K, V](),
		list: list,
	}
}

// Put saves or replaces a mapping
func (mp *KeyOrderedMapFunc[K, V]) Put(key K, value V) {
	length0 := mp.Map.Length()
	mp.Map.Put(key, value)
	if length0 < mp.Map.Length() {
		mp.list.Insert(key)
	}
}

// Delete removes mapping using key K.
//   - if key K is not mapped, the map is unchanged.
//   - O(log n)
func (m *KeyOrderedMapFunc[K, V]) Delete(key K) {
	m.Map.Delete(key)
	m.list.Delete(key)
}

// Clone returns a shallow clone of the map
func (m *KeyOrderedMapFunc[K, V]) Clear() {
	m.Map.Clear()
	m.list.Clear()
}

// Clone returns a shallow clone of the map
func (m *KeyOrderedMapFunc[K, V]) Clone() (clone *KeyOrderedMapFunc[K, V]) {
	return &KeyOrderedMapFunc[K, V]{Map: *m.Map.Clone(), list: m.list.Clone()}
}

// List provides keys in order
//   - O(n)
func (m *KeyOrderedMapFunc[K, V]) List(n ...int) (list []K) {
	return m.list.List(n...)
}
