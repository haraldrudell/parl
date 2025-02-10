/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package omaps

import "github.com/google/btree"

type orderedMapFunc[K comparable, V any] struct {
	OrderedMapFunc[K, V]
}

func newOrderedMapFunc[K comparable, V btree.Ordered](fieldp ...*orderedMapFunc[K, V]) (m *orderedMapFunc[K, V]) {

	// set m
	if len(fieldp) > 0 {
		m = fieldp[0]
	}
	if m == nil {
		m = &orderedMapFunc[K, V]{}
	}

	// initialize all fields
	var noLess func(a V, b V) (aBeforeB bool)
	NewOrderedMapFunc[K, V](noLess, &m.OrderedMapFunc)

	return
}

func newOrderedMapFuncUintptr[K comparable, V ~uintptr](less func(a V, b V) (aBeforeB bool), fieldp ...*orderedMapFunc[K, V]) (m *orderedMapFunc[K, V]) {

	// set m
	if len(fieldp) > 0 {
		m = fieldp[0]
	}
	if m == nil {
		m = &orderedMapFunc[K, V]{}
	}

	// initialize all fields
	NewOrderedMapFunc[K, V](less, &m.OrderedMapFunc)

	return
}
