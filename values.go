/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "iter"

// Values is a container for one or more values of type T
//   - used as the value for a map: map[int]parl.Values[string]
//   - implementation may be thread-safe by using atomics, lock or being read-only
//   - a similar value-type is AnyCount
//
// iteration:
//
//	for v := range values.Seq {
//	  v…
//	}
//
// about maps:
//   - to hold one or more values per key in a map value, a pointer is required
//   - the pointer can be implicitly declared as an interface value
//   - if Values0 implementation is not used, a value type for the map has to be defined locally
//   - the container-allocation for the first value is unavoidable
//     However, because Values0 contains a value field, the slice allocation
//     is deferred until the second value
//   - for values containing non-pointer atomics or locks, T must be pointer
type Values[T any] interface {
	// Add adds a value to the container
	Add(value T)
	// Count returns the number of values held in the container
	Count() (count int)
	// Seq allows for-range iteration over container values
	//	- for v := range values.Seq {
	Seq(yield func(value T) (keepGoing bool))
}

// Values.Seq is iter.Seq
var _ = func(v Values[int]) (seq iter.Seq[int]) { return v.Seq }
