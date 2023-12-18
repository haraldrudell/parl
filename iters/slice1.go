/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// Slice traverses a slice container. thread-safe
type Slice1[T any] struct {
	slice []T // the slice providing values
	// index is next slice-index to return
	//	- if index == len(slice) there are no more values
	index int
}

// NewSliceIterator returns an iterator iterating over slice T values
//   - thread-safe
//   - uses non-pointer atomics
func NewSlice1Iterator[T any](slice []T) (iterator Iterator[T]) { return &Slice[T]{slice: slice} }

// Init implements the right-hand side of a short variable declaration in
// the init statement for a Go “for” clause
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *Slice1[T]) Init() (iterationVariable T, iterator Iterator[T]) {
	iterator = i
	return
}

// Cond implements the condition statement of a Go “for” clause
//   - the iterationVariable is updated by being provided as a pointer.
//     iterationVariable cannot be nil
//   - errp is an optional error pointer receiving any errors during iterator execution
//   - condition is true if iterationVariable was assigned a value and the iteration should continue
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *Slice1[T]) Cond(iterationVariablep *T, errp ...*error) (condition bool) {
	if iterationVariablep == nil {
		cyclebreaker.NilError("iterationVariablep")
	}

	// check for next value
	var value T
	if value, condition = i.Next(); condition {
		*iterationVariablep = value
	}

	return // condition and iterationVariablep updated, errp unchanged
}

// Next advances to next item and returns it
//   - if hasValue true, value contains the next value
//   - otherwise, no more items exist and value is the data type zero-value
func (i *Slice1[T]) Next() (value T, hasValue bool) {
	var index = i.index
	if hasValue = index < len(i.slice); !hasValue {
		return // no more values
	}
	value = i.slice[index]
	i.index++

	return
}

// Cancel release resources for this iterator. Thread-safe
//   - not every iterator requires a Cancel invocation
func (i *Slice1[T]) Cancel(errp ...*error) (err error) {
	i.slice = nil
	return
}
