/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// InsertOrdered inserts value into an ordered slice
//   - duplicate values are allowed, new values are placed at the end
//   - insert is O(log n)
func InsertOrdered[E constraints.Ordered](slice0 []E, value E) (slice []E) {

	// find position
	var position int
	var wasFound bool
	if position, wasFound = slices.BinarySearch(slice0, value); wasFound {

		// advance beyond last identical value
		for {
			position++
			if position == len(slice0) || slice0[position] != value {
				break
			}
		}
	}

	return slices.Insert(slice0, position, value)
}

// InsertOrderedFunc inserts a value into a slice making it ordered using a comparison function.
//   - duplicate values are allowed, new values are placed at the end
//   - insert is O(log n)
//   - cmp function can be provided by E being a type with Cmp method.
//   - cmp(a, b) is expected to return an integer comparing the two parameters:
//     0 if a == b, a negative number if a < b and a positive number if a > b
func InsertOrderedFunc[E any](slice0 []E, value E, cmp func(a, b E) (result int)) (slice []E) {

	// obtain comparison function
	if cmp == nil {
		cmp = CmpFromComparable(value)
	}

	// find position
	var position int
	var wasFound bool
	if position, wasFound = slices.BinarySearchFunc(slice0, value, cmp); wasFound {

		// advance beyond last identical value
		for {
			position++
			if position == len(slice0) || cmp(value, slice0[position]) != 0 {
				break
			}
		}
	}

	return slices.Insert(slice0, position, value)
}

func CmpFromComparable[E any](value E) (cmp func(a, b E) (result int)) {

	// comparison function with value receiver
	if _, ok := any(value).(parl.Comparable[E]); ok {
		return func(a, b E) (result int) {
			return any(a).(parl.Comparable[E]).Cmp(b)
		}
	}

	// comparison function with pointer receiver
	if _, ok := any(&value).(parl.Comparable[E]); ok {
		return func(a, b E) (result int) {
			return any(&a).(parl.Comparable[E]).Cmp(b)
		}
	}

	panic(perrors.NewPF("cmp cannot be nil if value is not parl.Comparable"))
}
