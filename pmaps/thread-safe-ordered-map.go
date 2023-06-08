/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"github.com/haraldrudell/parl/pslices"
	"golang.org/x/exp/constraints"
)

// ThreadSafeOrderedMap is a mapping whose values are provided in order. Thread-safe.
type ThreadSafeOrderedMap[K comparable, V constraints.Ordered] struct {
	ThreadSafeOrderedMapFunc[K, V]
}

func NewThreadSafeOrderedMap[K comparable, V constraints.Ordered]() (orderedMap *ThreadSafeOrderedMap[K, V]) {
	return &ThreadSafeOrderedMap[K, V]{
		ThreadSafeOrderedMapFunc: *NewThreadSafeOrderedMapFunc2[K, V](
			pslices.NewOrdered[V](),
		)}
}

// Clone returns a shallow clone of the map
func (m *ThreadSafeOrderedMap[K, V]) Clone() (clone *ThreadSafeOrderedMap[K, V]) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return &ThreadSafeOrderedMap[K, V]{
		ThreadSafeOrderedMapFunc: *m.ThreadSafeOrderedMapFunc.Clone(),
	}
}
