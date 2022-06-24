/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "golang.org/x/exp/constraints"

// OrderedMap stores an ordered list of values accessible in constant time O(1).
// The map can be ordered by key value or by a custom key order.
// For large-sized structs, V can be pointer.
// Key must be constraints.Ordered, ie. not slice map or func or more.
//  pslices.NewOrderedMap
//  psclies.NewOrderedMapFunc
type OrderedMap[K constraints.Ordered, V any] interface {
	Get(key K, newV func() (value *V), makeV func() (value V)) (value V)
	Has(key K) (value V, ok bool)
	Delete(key K)
	List() (list []V)
}

// OrderedMap stores an ordered list of values accessible in constant time O(1).
// The map is in custom key order.
// For large-sized structs, V can be pointer.
// A function converts O to K
//  pslices.NewOrderedMapAny
type OrderedMapAny[O any, K constraints.Ordered, V any] interface {
	Get(key O, newV func() (value *V), makeV func() (value V)) (value V)
	Has(order O) (value V, ok bool)
	Delete(order O)
	List() (list []V)
}

// OrderedPointers[E any] is an ordered list of pointers sorted by the referenced values.
// OrderedPointers should be used when E is a large struct.
// pslices.NewOrderedPointers[E constraints.Ordered]() instantiates for pointers to
// comparable types, ie. not func slice map.
// pslices.NewOrderedAny instantiates for pointers to any type or for custom sort orders.
type OrderedPointers[E any] interface {
	Insert(element *E)
	List() (list []*E)
}

// OrderedValues[E any] is an ordered list of values sorted by value.
// OrderedValues should be used when E is an interface or a small-sized value.
// pslices.NewOrderedValues[E constraints.Ordered]() instantiates for
// comparable types, ie. not func map slice.
// pslices.NewOrderedAny[E any](cmp func(a, b E) (result int)) instantiates for any type
// or for custom sort order.
type OrderedValues[E any] interface {
	Insert(element E)
	Delete(element E)
	List() (list []E)
}
