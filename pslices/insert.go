/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// InsertOrdered inserts values into a slice making it ordered
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

// InsertOrderedFunc inserts values into a slice making it ordered using a comparison function
func InsertOrderedFunc[E any](slice0 []E, value E, cmp func(E, E) (result int)) (slice []E) {
	if cmp == nil {
		var c Comparable[E]
		var ok bool
		if c, ok = any(value).(Comparable[E]); ok {
			cmp = c.Cmp
		}
		if cmp == nil {
			panic(perrors.New("InsertOrderedFunc with cmp nil"))
		}
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
