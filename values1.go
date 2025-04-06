/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Values1 is thread-safe Values
type Values1[T any] struct {
	// vLock makes v access thread-safe
	vLock Mutex
	// wrapped non-thread-safe container
	values0 Values0[T]
}

// Values1 implements the Values interface
var _ Values[int] = &Values1[int]{}

// NewValuesThreadSafe returns a [Values] interface any-count map-value implementation
//   - values: zero or more values put into the container
//   - Values0 is initialization-free
//   - —
//   - the map is map[K]V where V is [parl.Value[T]]
//   - because V is interface, it is a pointer copied by the map
//   - the value of a Go map must be on the heap and it is allocated here
func NewValuesThreadSafe[T any](values ...T) (v Values[T]) {

	// allocate container on the heap
	v = &Values1[T]{}

	// put possible values into container
	v.AddPreAlloc(PreAllocYes, values...)

	return
}

// Add adds a value to the container
//   - pre-alloc 10 elements
//   - Add does not receive values-slice ownership,
//     slice allocation may result
func (v *Values1[T]) Add(values ...T) { v.AddPreAlloc(allocTen, values...) }

// AddPreAlloc adds values to the container with slice pre-allocation
//   - size: number of elements pre-allocation
//   - size [PreAllocYes]: pre-alloc 4 Kib or 10 elements
//   - size [PreAllocLow]: pre-alloc 10 elements
//   - AddPreAlloc does not receive values-slice ownership,
//     slice allocation may result
func (v *Values1[T]) AddPreAlloc(size int, values ...T) {
	defer v.vLock.Lock().Unlock()

	v.values0.AddPreAlloc(size, values...)
}

// SetValues empties the container and uses values
//   - container takes ownership of values slice
//   - allocation-free assignment of multiple values
//   - values nil: de-allocates
func (v *Values1[T]) SetValues(values []T) {
	defer v.vLock.Lock().Unlock()

	v.values0.SetValues(values)
}

// Count returns the number of values held in the container
func (v *Values1[T]) Count() (count int) {
	defer v.vLock.Lock().Unlock()

	return v.values0.Count()
}

// Seq allows for-range iteration over container values
//   - for v := range values.Seq {
func (v *Values1[T]) Seq(yield func(value T) (keepGoing bool)) {

	var i int
	for {

		// fetch and provide value
		var value, hasValue = v.getValue(i)
		if !hasValue {
			return
		} else if !yield(value) {
			return
		}

		i++
	}
}

// getValue fetches value by index
func (v *Values1[T]) getValue(index int) (value T, hasValue bool) {
	defer v.vLock.Lock().Unlock()

	return v.values0.GetN(index)
}
