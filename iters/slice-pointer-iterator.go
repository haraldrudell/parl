/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

// SlicePointerIterator traverses a slice container using pointers to value. thread-safe.
//   - the difference is that:
//   - instead of copying a value from the slice,
//   - a pointer to the slice value is returned
type SlicePointerIterator[E any] struct {
	slice []E // the slice providing values

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

	// Delegator implements the value methods required by the [Iterator] interface
	//   - Next HasNext NextValue
	//     Same Has SameValue
	//   - the delegate provides DelegateAction[T] function
	Delegator[*E]
}

// NewSlicePointerIterator returns an iterator of pointers to T
//   - the difference is that:
//   - instead of copying a value from the slice,
//   - a pointer to the slice value is returned
//   - the returned [Iterator] value cannot be copied, the pointer value
//     must be used
//   - uses self-referencing pointers
func NewSlicePointerIterator[E any](slice []E) (iterator Iterator[*E]) {
	i := SlicePointerIterator[E]{slice: slice}
	i.Delegator = *NewDelegator(i.delegateAction)
	return &i
}

func NewSlicePointerIteratorField[E any](fieldp *SlicePointerIterator[E], slice []E) (iterator Iterator[*E]) {
	fieldp.slice = slice
	fieldp.Delegator = *NewDelegator(fieldp.delegateAction)
	iterator = fieldp
	return
}

// Init implements the right-hand side of a short variable declaration in
// the init statement for a Go “for” clause
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *SlicePointerIterator[E]) Init() (iterationVariable *E, iterator Iterator[*E]) {
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
func (i *SlicePointerIterator[E]) Cond(iterationVariablep **E, errp ...*error) (condition bool) {
	if iterationVariablep == nil {
		perrors.NewPF("iterationVariablep cannot bee nil")
	}

	// check for next value
	var value *E
	if value, condition = i.delegateAction(IsNext); condition {
		*iterationVariablep = value
	}

	return // condition and iterationVariablep updated, errp unchanged
}

// Cancel release resources for this iterator. Thread-safe
//   - not every iterator requires a Cancel invocation
func (i *SlicePointerIterator[E]) Cancel(errp ...*error) (err error) {
	i.isEnd.CompareAndSwap(false, true)
	return
}

// Next finds the next or the same value. Thread-safe
//   - isSame true means first or same value should be returned
//   - value is the sought value or the T type’s zero-value if no value exists
//   - hasValue true means value was assigned a valid T value
func (i *SlicePointerIterator[E]) delegateAction(isSame NextAction) (value *E, hasValue bool) {

	if i.isEnd.Load() {
		return // no more values return
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	if i.isEnd.Load() {
		return // no more values return
	}

	// for IsSame operation the first value must be sought
	//	- therefore, if the first value has not been sought, seek it now or
	//	- if not IsSame operation, advance to the next value
	if !i.didNext || isSame != IsSame {

		// note that first value has been sought
		if !i.didNext {
			i.didNext = true
		}

		// find slice index to use
		//	- if a value was found, advance index
		//	- final i.index value is len(i.slice)
		if i.hasValue {
			i.index++
		}

		// check if the new index is within available slice values
		//	- when i.index has reached len(i.slice), i.hasValue is always false
		//	- when hasValue is false, i.index will no longer be incremented
		i.hasValue = i.index < len(i.slice)
	}

	// update hasValue and value
	//	- get the value if it is valid, otherwise zero-value
	if hasValue = i.hasValue; hasValue {
		value = &i.slice[i.index]
	} else {
		i.isEnd.CompareAndSwap(false, true)
	}

	return // value and hasValue indicates availability
}
