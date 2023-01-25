/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"fmt"
	"sync"

	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
)

// type Elements[T comparable] []T

// ElementSlice ToDo
type ElementSlice[T comparable, E any] struct {
	lock     sync.Mutex
	didNext  bool // indicates whether any value has been sought
	hasValue bool // indicates whether index has been verified to be valid
	index    int  // index in slice, 0…len(slice)
	slice    []E
}

// NewElements returns an iterator of interface-type sets.Element[T] based from a
// slice of non-interface-type Elements[T comparable].
func NewElements[T comparable, E any](elements []E) (iter iters.Iterator[Element[T]]) {

	// runtime check if requyired type conversion works
	var e *E
	var a any = e
	if _, ok := a.(Element[T]); !ok {
		var returnElementTypeStr string
		var t *Element[T]
		if returnElementTypeStr = fmt.Sprintf("%T", t); len(returnElementTypeStr) > 0 {
			returnElementTypeStr = returnElementTypeStr[1:]
		}
		var eTypeStr string
		if eTypeStr = fmt.Sprintf("%T", t); len(eTypeStr) > 0 {
			eTypeStr = eTypeStr[1:]
		}
		panic(perrors.ErrorfPF("input type %s does not implement interface-type %s",
			eTypeStr, returnElementTypeStr,
		))
	}

	// create the iterator
	slice := ElementSlice[T, E]{slice: elements}
	return &iters.Delegator[Element[T]]{Delegate: &slice}
}

func (iter *ElementSlice[T, E]) Next(isSame iters.NextAction) (value Element[T], hasValue bool) {
	iter.lock.Lock()
	defer iter.lock.Unlock()

	// if next operation has not completed, we do not know if a value exist,
	// and next operation must be completed.
	// if next has completed and we seek the same value, next operation should not be done.
	if !iter.didNext || isSame != iters.IsSame {

		// find slice index to use
		if iter.hasValue {
			// if a value has been found and is valid, advance index.
			// the final value for iter.index is len(iter.slice)
			iter.index++
		}

		// check if the new index is within available slice values
		// when iter.index has reached len(iter.slice), iter.hasValue is always false.
		// when hasValue is false, iter.index will no longer be incremented.
		iter.hasValue = iter.index < len(iter.slice)

		// indicate that iter.hasValue is now valid
		if !iter.didNext {
			iter.didNext = true
		}
	}

	// get the value if it is valid, otherwise zero-value
	if hasValue = iter.hasValue; hasValue {
		var ePointer *E = &iter.slice[iter.index]
		var a any = ePointer
		var ok bool
		if value, ok = a.(Element[T]); !ok {
			// this is checked in NewElements: should never happen
			panic(perrors.ErrorfPF("type assertion failed: %T %T", ePointer, value))
		}
	}

	return // value and hasValue indicates availability
}

func (iter *ElementSlice[T, E]) Cancel() (err error) {
	iter.lock.Lock()
	defer iter.lock.Unlock()

	iter.hasValue = false // invalidate iter.value
	iter.slice = nil      // prevent any next operation
	return
}
