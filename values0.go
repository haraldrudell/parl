/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Values0 discussion
//	- because we want a map key mapping to one or more values,
//		the map declaration value type must be pointer
//	- for flexibility, a pointer may be implicitly declared as an interface value
//	- the interface is implemented by a pointer to a concrete type
//	- the concrete type cannot be declared as []T
//	- on append, the value of the slice []T may change
//	- therefore, Values0 must be struct with a field []T
//	- to avoid immediate slice-allocation, a value field T is added
//	- to allow T any value, a hasValue flag is required to indicate if T is present

// Values0 is a value container for any number of values
type Values0[T any] struct {
	// value recognized when hasValue is true
	//	- precedes values during iteration
	value T
	// values are present regardless of hasValue
	values []T
	// true: value is present
	hasValue bool
}

// Values0 implements the Values interface
var _ Values[int] = &Values0[int]{}

// NewValues returns a [Values] interface any-count map-value implementation
//   - values: zeero or more values put into the container
//   - —
//   - the map is map[K]V where V is [parl.Value[T]]
//   - because V is interface, it is a pointer copied by the map
//   - the value pointed to must be on the heap and it is allocated here
func NewValues[T any](values ...T) (v Values[T]) {

	// allocate any-count value container on the heap
	v = &Values0[T]{}

	// put values into container
	for _, t := range values {
		v.Add(t)
	}

	return
}

// Add adds a value to the container
func (v *Values0[T]) Add(value T) {

	// try value
	if !v.hasValue {
		v.value = value
		v.hasValue = true
		return // value set return
	}

	// check for slice
	if len(v.values) == 0 {
		v.values = []T{value}
		return
	} // initial slice-allocation return

	// slice append
	v.values = append(v.values, value)
}

// Count returns the number of values held in the container
func (v *Values0[T]) Count() (count int) {
	count = len(v.values)
	if v.hasValue {
		count++
	}
	return
}

// Seq allows for-range iteration over container values
//   - for v := range values.Seq {
func (v *Values0[T]) Seq(yield func(value T) (keepGoing bool)) {
	var i int
	for {
		if i == 0 && v.hasValue {
			i++
			if !yield(v.value) {
				return
			}
		}
		var sliceIndex = i
		if v.hasValue {
			sliceIndex--
		}
		if sliceIndex >= len(v.values) {
			break
		}
		i++
		if !yield(v.values[sliceIndex]) {
			return
		}
	}
}
