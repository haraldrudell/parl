/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"errors"
	"sync"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

const (
	FunctionIteratorCancel int = -1
)

// FunctionIterator traverses a function generatoing values. thread-safe and re-entrant.
type FunctionIterator[T any] struct {
	lock sync.Mutex
	// didNext indicates that a Next operation has completed and that hasValue may be valid.
	didNext bool
	// index contains the next index to use for fn invocations
	index int
	// fn will be invoked with an index 0… until fn returns an error.
	// if fn returns error, it will not be invoked again.
	// fn signals end of values by returning parl.ErrEndCallbacks.
	// when fn returns parl.ErrEndCallbacks, the accompanying value is not used.
	// if index == FunctionIteratorCancel, it means Cancel.
	// It is the last invocation of fn and fn should release any resources.
	// fn is typically expected to be re-entrant and thread-safe.
	fn func(index int) (value T, err error)
	// value is a returned value from fn.
	// value is valid when isEnd is false and hasValue is true.
	value T
	// err contains any errors returned by fn other than parl.ErrEndCallbacks
	err error
	// isEnd indicates that fn returned error or Cancel was invoked.
	// isEnd invalidates both value and hasValue.
	isEnd bool
	// hasValue indicates that value is valid.
	hasValue bool
}

// NewFunctionIterator returns a parli.Iterator based on a function.
// functionIterator is thread-safe and re-entrant.
func NewFunctionIterator[T any](
	fn func(index int) (value T, err error),
) (iterator Iterator[T]) {
	if fn == nil {
		panic(perrors.NewPF("fn cannot be nil"))
	}
	return &Delegator[T]{Delegate: &FunctionIterator[T]{fn: fn}}
}

// InitFunctionIterator initializes a FunctionIterator struct.
// functionIterator is thread-safe and re-entrant.
func InitFunctionIterator[T any](iterp *FunctionIterator[T], fn func(index int) (value T, err error)) {
	if iterp == nil {
		panic(perrors.NewPF("iterator cannot be nil"))
	}
	if fn == nil {
		panic(perrors.NewPF("fn cannot be nil"))
	}
	iterp.lock = sync.Mutex{}
	iterp.didNext = false
	iterp.index = 0
	iterp.fn = fn
	var value T
	iterp.value = value
	iterp.err = nil
	iterp.isEnd = false
	iterp.hasValue = false
}

func (iter *FunctionIterator[T]) Next(isSame NextAction) (value T, hasValue bool) {
	iter.lock.Lock()
	defer iter.lock.Unlock()

	// should next operation be invoked?
	if iter.isEnd {
		return // error from fn or cancel invoked: zero-value and hasValue false
	}

	// same value when a value has previously been read
	if isSame == IsSame && iter.didNext {
		value = iter.value
		hasValue = iter.hasValue
		return
	} else if !iter.didNext {
		iter.didNext = true
	}

	// invoke fn
	index := iter.index // remember out index
	iter.index++        // update index for next invocation
	var err error
	// invokeFn unlocks while invoking fn to allow re-entrancy.
	// if fn returns error, invokeFn updates: iter.isEnd iter.err iter.value iter.hasValue.
	if value, hasValue, err = iter.invokeFn(index); err != nil {
		return // fn error: zero-value, false
	}

	// update state for subsequent Same invocations
	if index+1 == iter.index { // there were no intermediate invocations
		iter.value = value
		iter.hasValue = hasValue
	}

	return // good return: value and hasValue both valid
}

// invokeFn invokes iter.fn unlocked, recovers from fn panics and updates iter.err.
func (iter *FunctionIterator[T]) invokeFn(index int) (value T, hasValue bool, err error) {

	// invoke fn with unlock and panic recovery
	cyclebreaker.RecoverInvocationPanic(func() {
		// it is allowed for fn to invoke the iterator
		iter.lock.Unlock()
		defer iter.lock.Lock()

		value, err = iter.fn(index)
	}, &err)

	// update error outcome
	if err != nil {
		iter.isEnd = true
		if !errors.Is(err, cyclebreaker.ErrEndCallbacks) {
			iter.err = perrors.AppendError(iter.err, err)
		}
		var zeroValue T
		value = zeroValue
		iter.value = zeroValue
		iter.hasValue = false
	} else {
		hasValue = true
	}

	return
}

func (iter *FunctionIterator[T]) Cancel() (err error) {
	iter.lock.Lock()
	defer iter.lock.Unlock()

	// if already error, or Cancel has alredy been invoked, return same result
	if iter.isEnd {
		err = iter.err
		return
	}
	iter.isEnd = true

	// execute cancel
	_, _, err = iter.invokeFn(FunctionIteratorCancel)

	return
}
