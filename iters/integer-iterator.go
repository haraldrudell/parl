/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"golang.org/x/exp/constraints"
)

type IntegerIterator[T constraints.Integer] struct {
	lastValue, delta T
	// isEnd is fast outside-lock check for no values available
	isEnd atomic.Bool

	lock sync.Mutex
	//	- behind lock
	didReturnAValue bool
	//	- behind lock
	value T
}

func NewIntegerIterator[T constraints.Integer](firstValue, lastValue T) (iterator Iterator[T]) {
	i := &IntegerIterator[T]{value: firstValue, lastValue: lastValue}
	if firstValue > lastValue {
		var i64 = int64(-1)
		i.delta = T(i64)
	} else {
		i.delta = T(1)
	}
	return i
}

// Init implements the right-hand side of a short variable declaration in
// the init statement of a Go “for” clause
//
//		for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *IntegerIterator[T]) Init() (iterationVariable T, iterator Iterator[T]) { iterator = i; return }

// Cond implements the condition statement of a Go “for” clause
//   - condition is true if iterationVariable was assigned a value and the iteration should continue
//   - the iterationVariable is updated by being provided as a pointer.
//     iterationVariable cannot be nil
//   - errp is an optional error pointer receiving any errors during iterator execution
//
// Usage:
//
//	for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	  // i is pointer to slice element
func (i *IntegerIterator[T]) Cond(iterationVariablep *T, errp ...*error) (condition bool) {
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
func (i *IntegerIterator[T]) Next() (value T, hasValue bool) { return i.nextSame(IsNext) }

// Same returns the same value again
//   - if hasValue true, value is valid
//   - otherwise, no more items exist and value is the data type zero-value
//   - If Next or Cond has not been invoked, Same first advances to the first item
func (i *IntegerIterator[T]) Same() (value T, hasValue bool) { return i.nextSame(IsSame) }

// Cancel stops an iteration
//   - after Cancel invocation, Cond, Next and Same indicate no value available
//   - Cancel returns the first error that occurred during iteration, if any
//   - an iterator implementation may require Cancel invocation
//     to release resources
//   - Cancel is deferrable
func (i *IntegerIterator[T]) Cancel(errp ...*error) (err error) {
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
func (i *IntegerIterator[T]) nextSame(isSame NextAction) (value T, hasValue bool) {

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

	// note that first value has been sought
	if !i.didReturnAValue {
		i.didReturnAValue = true
	} else if isSame == IsNext {
		if i.value == i.lastValue {
			if !i.isEnd.Load() {
				i.isEnd.CompareAndSwap(false, true)
			}
			return
		}
		i.value += i.delta
	}

	value = i.value
	hasValue = true

	return // value and hasValue valid
}
