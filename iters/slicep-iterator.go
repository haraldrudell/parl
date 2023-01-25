/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

// SlicePIterator traverses a slice container using pointers to value. thread-safe.
package iters

import (
	"sync"

	"github.com/haraldrudell/parl/perrors"
)

// SlicePIterator traverses a slice container using pointers to value. thread-safe.
type SlicePIterator[T any] struct {
	lock     sync.Mutex
	didNext  bool // indicates whether any value has been sought
	hasValue bool // indicates whether index has been verified to be valid
	index    int  // index in slice, 0…len(slice)
	slice    []T
}

// NewSlicePIterator returns an iterator of pointers to T.
// SlicePIterator is thread-safe.
func NewSlicePIterator[T any](slice []T) (iterator Iterator[*T]) {
	return &Delegator[*T]{Delegate: &SlicePIterator[T]{slice: slice}}
}

// InitSliceIterator initializes a SliceIterator struct.
// sliceIterator is thread-safe.
func InitSlicePIterator[T any](iterp *SliceIterator[T], slice []T) {
	if iterp == nil {
		panic(perrors.NewPF("iterator cannot be nil"))
	}
	iterp.lock = sync.Mutex{}
	iterp.didNext = false
	iterp.hasValue = false
	iterp.index = 0
	iterp.slice = slice
}

func (iter *SlicePIterator[T]) Next(isSame NextAction) (value *T, hasValue bool) {
	iter.lock.Lock()
	defer iter.lock.Unlock()

	// if next operation has not completed, we do not know if a value exist,
	// and next operation must be completed.
	// if next has completed and we seek the same value, next operation should not be done.
	if !iter.didNext || isSame != IsSame {

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
		value = &iter.slice[iter.index]
	}

	return // value and hasValue indicates availability
}

func (iter *SlicePIterator[T]) Cancel() (err error) {
	iter.lock.Lock()
	defer iter.lock.Unlock()

	iter.hasValue = false // invalidate iter.value
	iter.slice = nil      // prevent any next operation
	return
}
