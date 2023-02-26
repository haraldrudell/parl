/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// KeyOrderedMapFunc is a mapping whose keys are provided in custom order.
package pmaps

import (
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// KeyOrderedMapFunc is a mapping whose keys are provided in custom order.
type KeyOrderedMapFunc[K constraints.Ordered, V any] struct {
	KeyOrderedMap[K, V]
}

// NewKeyOrderedMapFunc returns a mapping whose keys are provided in custom order.
func NewKeyOrderedMapFunc[K constraints.Ordered, V any](
	cmp func(a, b K) (result int),
) (orderedMap *KeyOrderedMapFunc[K, V]) {
	return &KeyOrderedMapFunc[K, V]{
		KeyOrderedMap: *newKeyOrderedMap[K, V](pslices.NewOrderedAny(cmp)),
	}
}
