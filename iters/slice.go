/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// Slice traverses a slice container. thread-safe
type Slice[T any] struct {
	slice []T // the slice providing values

	// isEnd is fast outside-lock check for no values available
	isEnd atomic.Bool

	// lock serializes Next and Cancel invocations
	lock sync.Mutex
	// didNext indicates that the first value was sought
	//	- behind lock
	didNext bool
	// hasValue indicates that slice[index] is the current value
	//	- behind lock
	hasValue bool
	// index in slice, 0…len(slice)
	//	- behind lock
	index int
}

// NewSliceIterator returns an iterator iterating over slice T values
//   - thread-safe
func NewSliceIterator[T any](slice []T) (iterator Iterator[T]) {
	return &Slice[T]{slice: slice}
}

// Init implements the right-hand side of a short variable declaration in
// the init statement for a Go “for” clause
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *Slice[T]) Init() (iterationVariable T, iterator Iterator[T]) {
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
func (i *Slice[T]) Cond(iterationVariablep *T, errp ...*error) (condition bool) {
	if iterationVariablep == nil {
		cyclebreaker.NilError("iterationVariablep")
	}

	// check for next value
	var value T
	if value, condition = i.nextSame(IsNext); condition {
		*iterationVariablep = value
	}

	return // condition and iterationVariablep updated, errp unchanged
}

// Next advances to next item and returns it
//   - if hasValue true, value contains the next value
//   - otherwise, no more items exist and value is the data type zero-value
func (i *Slice[T]) Next() (value T, hasValue bool) { return i.nextSame(IsNext) }

// Same returns the same value again
//   - if hasValue true, value is valid
//   - otherwise, no more items exist and value is the data type zero-value
//   - If Next or Cond has not been invoked, Same first advances to the first item
func (i *Slice[T]) Same() (value T, hasValue bool) { return i.nextSame(IsSame) }

// Cancel release resources for this iterator. Thread-safe
//   - not every iterator requires a Cancel invocation
func (i *Slice[T]) Cancel(errp ...*error) (err error) {
	if i.isEnd.Load() {
		return // already canceled
	}
	i.isEnd.CompareAndSwap(false, true)

	return
}

// nextSame finds the next or the same value. Thread-safe
//   - isSame == IsSame means first or same value should be returned
//   - value is the sought value or the T type’s zero-value if no value exists
//   - hasValue true means value was assigned a valid T value
func (i *Slice[T]) nextSame(isSame NextAction) (value T, hasValue bool) {

	// outside lock check
	if i.isEnd.Load() {
		return // no more values return
	}
	i.lock.Lock()
	defer i.lock.Unlock()

	// inside lock check
	if i.isEnd.Load() {
		return // no more values return
	}

	// IsNext always seeks the next value
	//	- IsSame prior to any value returned also seeks the next value
	if !i.didNext || isSame == IsNext {

		// note that first value has been sought
		if !i.didNext {
			i.didNext = true
		}

		// find slice index to use
		//	- if a value was previously returned, advance index
		//	- final i.index value is len(i.slice)
		if i.hasValue {
			i.index++
		}

		// check if the new index is within available slice values
		//	- when i.index has reached len(i.slice), i.hasValue is always false
		//	- when hasValue is false, i.index will no longer be incremented
		i.hasValue = i.index < len(i.slice)
	}

	// update hasValue, value and i.isEnd
	//	- get the value if it is valid, otherwise zero-value
	if hasValue = i.hasValue; hasValue {
		value = i.slice[i.index]
	} else if !i.isEnd.Load() {
		i.isEnd.CompareAndSwap(false, true)
	}

	return // value and hasValue valid
}
