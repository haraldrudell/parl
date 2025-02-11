/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "iter"

// AnyCount is a value type that may contain
// zero, one or many values
//   - allows a single value to represent any number of values
//   - zero or one value requires no allocation
//   - — taking address of AnyCount causes allocation
//   - slice allocation is only required for more than one value
//   - allocation-free iteration [AnyCount.Init] [AnyCount.Cond]
//   - [AnyCount.Count] counts contained values
//   - Count, iteration, or update are not thread-safe
//   - if the contained value includes non-pointer locks or atomics, T must be pointer to value
//   - has no new function
//   - a similar type with pointer receiver is [Values] and [NewValues]
//
// Iteration, allocation-free:
//
//	var a AnyCount[T]
//	for i, v := a.Init(); a.Cond(&i, &v); {
//		v…
type AnyCount[T any] struct {
	// value recognized when hasValue is true
	//	- precedes Values during iteration
	Value T
	// Values are present regardless of HasValue
	Values []T
	// true: Value is present
	HasValue bool
}

var _ iter.Seq[int] = (&AnyCount[int]{}).Seq

// Count returns number of values
func (a AnyCount[T]) Count() (count int) {
	count = len(a.Values)
	if a.HasValue {
		count++
	}
	return
}

// Seq allows for-range iteration over container values
//   - for v := range values.Seq {
func (a AnyCount[T]) Seq(yield func(value T) (keepGoing bool)) {
	var i int
	for {
		if i == 0 && a.HasValue {
			i++
			if !yield(a.Value) {
				return
			}
		}
		var sliceIndex = i
		if a.HasValue {
			sliceIndex--
		}
		if sliceIndex >= len(a.Values) {
			break
		}
		i++
		if !yield(a.Values[sliceIndex]) {
			return
		}
	}
}
