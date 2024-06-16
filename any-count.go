/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

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
//
// Iteration, allocation-free:
//
//	var a AnyCount[T]
//	for i, v := a.Init(); a.Cond(&i, &v); {
//		v…
type AnyCount[T any] struct {
	// value is present when hasValue is true
	Value T
	// Values are present regardless of HasValue
	Values []T
	// true: Value is present
	HasValue bool
}

// Count returns number of values
func (a AnyCount[T]) Count() (count int) {
	count = len(a.Values)
	if a.HasValue {
		count++
	}
	return
}

// Init implements the right-hand side of a short variable declaration in
// the init statement of a Go “for” clause
//   - iteration is not thread-safe
func (a AnyCount[T]) Init() (firstIndex int, value T) {
	if a.HasValue {
		firstIndex = -1
	}
	return
}

// Cond is the condition statement of a Go “for” clause
//   - iteration is not thread-safe
func (a AnyCount[T]) Cond(indexp *int, valuep *T) (condition bool) {
	var i = *indexp

	// iterate Value
	if condition = i == -1; condition {
		*valuep = a.Value
		*indexp = 0
		return
	}

	// iterate Values
	if condition = i < len(a.Values); !condition {
		return // end of values
	}
	*valuep = a.Values[i]
	*indexp++

	return
}
