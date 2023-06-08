/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyOrderedMap is a mapping whose keys are provided in order.
package pmaps

import (
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// KeyOrderedMap is a mapping whose keys are provided in order.
//   - key is ordered type
type KeyOrderedMap[K constraints.Ordered, V any] struct {
	KeyOrderedMapFunc[K, V]
}

// NewKeyOrderedMap returns a mapping whose keys are provided in order.
func NewKeyOrderedMap[K constraints.Ordered, V any]() (orderedMap *KeyOrderedMap[K, V]) {
	return &KeyOrderedMap[K, V]{
		KeyOrderedMapFunc: *NewKeyOrderedMapFunc2[K, V](
			pslices.NewOrdered[K](),
		),
	}
}

// Clone returns a shallow clone of the map
func (m *KeyOrderedMap[K, V]) Clone() (clone *KeyOrderedMap[K, V]) {
	return &KeyOrderedMap[K, V]{
		KeyOrderedMapFunc: *m.KeyOrderedMapFunc.Clone(),
	}
}
