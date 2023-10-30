/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import "github.com/google/btree"

type orderedMapFunc[K comparable, V any] struct {
	OrderedMapFunc[K, V]
}

func newOrderedMapFunc[K comparable, V btree.Ordered]() (m *orderedMapFunc[K, V]) {
	return &orderedMapFunc[K, V]{
		OrderedMapFunc: *NewOrderedMapFunc[K, V](nil),
	}
}

func newOrderedMapFuncUintptr[K comparable, V ~uintptr](less func(a V, b V) (aBeforeB bool)) (m *orderedMapFunc[K, V]) {
	return &orderedMapFunc[K, V]{
		OrderedMapFunc: *NewOrderedMapFunc[K, V](less),
	}
}
