/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "unsafe"

const (
	// [Values0.AddPreAlloc] efficient pre-allocation
	PreAllocYes = 0
	// [Values0.AddPreAlloc] low pre-allocation
	PreAllocLow = 10
	// [Values0.AddPreAlloc] no pre-allocation
	NoPreAlloc = 1
)

// Values0 discussion
//	- for a Go map mapping to be updatable without rewriting the mapping,
//		the map declaration value type must be pointer
//	- for flexibility, a pointer may be implicitly declared as an interface value
//	- the interface is implemented by a pointer to a concrete type
//	- the concrete type cannot be declared as []T,
//		because on append, the value of the slice []T may change
//	- therefore, Values0 must be struct with a field []T
//	- to avoid immediate slice-allocation, an allocation-free value field T is added
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
//   - values: zero or more values put into the container
//   - Values0 is initialization-free
//   - —
//   - the map is map[K]V where V is [parl.Value[T]]
//   - because V is interface, it is a pointer copied by the map
//   - the value of a Go map must be on the heap and it is allocated here
func NewValues[T any](values ...T) (v Values[T]) {

	// allocate container on the heap
	v = &Values0[T]{}

	// put possible values into container
	v.AddPreAlloc(PreAllocYes, values...)

	return
}

// Add adds a value to the container
//   - pre-alloc 10 elements
//   - Add does not receive values-slice ownership,
//     slice allocation may result
func (v *Values0[T]) Add(values ...T) { v.AddPreAlloc(allocTen, values...) }

// AddPreAlloc adds values to the container with slice pre-allocation
//   - size: number of elements pre-allocation
//   - size [PreAllocYes]: pre-alloc 4 Kib or 10 elements
//   - size [PreAllocLow]: pre-alloc 10 elements
//   - AddPreAlloc does not receive values-slice ownership,
//     slice allocation may result
func (v *Values0[T]) AddPreAlloc(size int, values ...T) {

	// noop check
	if len(values) == 0 {
		return
	}
	// values is not empty

	// try value
	if !v.hasValue && len(v.values) == 0 {
		v.value = values[0]
		v.hasValue = true
		if len(values) == 1 {
			return // single value set return
		}
		values = values[1:]
	}

	// check for existing slice
	if cap(v.values) > 0 {
		v.values = append(v.values, values...)
		return // appended to existing slice return
	}
	// v.values must be allocated

	// determine 4 KiB or 10 elements
	if size <= PreAllocYes {
		var sizeT = int(unsafe.Sizeof(v.value))
		size = max(alloc4Kib/sizeT, allocTen)
	}

	// ensure values to fit
	if size < len(values) {
		size = len(values)
	}

	// make, copy, save
	var slice = make([]T, len(values), size)
	copy(slice, values)
	v.values = slice
}

// SetValues empties the container and uses values
//   - container takes ownership of values slice
//   - allocation-free assignment of multiple values
//   - values nil: de-allocates
func (v *Values0[T]) SetValues(values []T) {

	// zero-out
	if v.hasValue {
		var t T
		v.value = t
		v.hasValue = false
	}

	v.values = values
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

	// v.value
	if v.hasValue {
		if !yield(v.value) {
			return
		}
	}

	// v.values
	for _, v := range v.values {
		if !yield(v) {
			return
		}
	}
}

// GetN retrives values by index
func (v *Values0[T]) GetN(index int) (value T, hasValue bool) {

	// invalid index
	if index < 0 {
		return
	}
	// index >= 0

	if v.hasValue {
		if index == 0 {
			value = v.value
			hasValue = true
			return
		}
		index--
	}
	// index >= 0 is index in v.values

	if index >= len(v.values) {
		return // too large index
	}

	value = v.values[index]
	hasValue = true

	return
}

const (
	// 4 KiB bytes
	alloc4Kib = 4096
	// 10 elements
	allocTen = 10
)
