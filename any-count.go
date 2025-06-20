/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "iter"

// AnyCount is a tuple: an initialization-free value-type container of
// zero, one or many values
//   - purpose is a non-thread-safe function parameter or result
//   - — zero-allocation for zero or one values
//   - — easy one-go iteration over zero/one/many values: [AnyCount.Seq]
//   - thread-safe heap-allocated alternative usable for heap entities like
//     map value, slice element or atomic pointer: [Values] and [NewValues]
//   - —
//   - allows a single value to represent any number of values
//   - taking address of AnyCount may cause allocation
//   - [AnyCount.Count] counts contained values
//   - Count, iteration, or update are not thread-safe
//   - if contained type includes non-pointer locks or atomics,
//     T must be pointer to value
//
// Iteration, allocation-free:
//
//	var a AnyCount[T]
//	for v := range a.Seq {
//		v…
type AnyCount[T any] struct {
	// value recognized when hasValue is true
	//	- precedes Values during iteration
	//	- in iteration order, the first value if present
	Value T
	// Values are present regardless of HasValue
	Values []T
	// true: Value is present
	HasValue bool
}

// AnyCount is [iter.Seq]
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

	// Value first
	if a.HasValue {
		if !yield(a.Value) {
			return
		}
	}

	// Values
	for _, v := range a.Values {
		if !yield(v) {
			return
		}
	}
}
