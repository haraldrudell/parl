/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plog"
)

const (
	// invoke iteratorFunction for next value
	enqueueForFnNextItem = false
	// invoke iteratorFunction to request cancel
	enqueueForFnCancel = true
)

// baseIterator is a pointed-to enclosed type due to BaseIterator providing
// baseIterator.delegateAction
type baseIterator[T any] struct {
	invokeFn InvokeFunc[T]

	// publicsLock serializes invocations of i.Next and i.Cancel
	publicsLock sync.Mutex

	// Next

	// didNext indicates that a Next operation has completed and that hasValue may be valid
	//	- behind publicsLock
	didNext bool
	// value is a returned value from fn.
	// value is valid when isEnd is false and hasValue is true.
	//	- behind publicsLock
	value T
	// hasValue indicates that value is valid.
	//	- behind publicsLock
	hasValue bool

	// indicates that no further values can be returned.
	// caused by:
	//	- error received from iterator
	//	- iterator cancel completed
	//	- iterator signaled end of values
	//	- —
	//	- written behind publicsLock
	noValuesAvailable atomic.Bool

	// enqueueForFn

	// cancelState is updated by:
	//	- Cancel() and
	//	- enqueueForFn() inside i.publicsLock lock
	cancelState atomic.Uint32
	// errWait waits until
	//	- an error is received from fn
	//	- cancel returns from fn FunctionIteratorCancel -1 invocation
	//		without error
	//	- fn signals end of data ny returning parl.ErrEndCallbacks
	//	- updated by enqueueForFn() inside i.publicsLock lock
	errWait sync.WaitGroup
	// err contains any errors returned by fn other than parl.ErrEndCallbacks
	//	- nil if fn is active
	//	- non-nil pointer to nil error if fn completed without error
	//	- — returned parl.ErrEndCallbacks
	//	- — returned from FunctionIteratorCancel -1 invocation without error
	//	- non-nil error if fn returned error
	//	- updated by enqueueForFn() inside i.publicsLock lock
	err atomic.Pointer[error]
}

// BaseIterator implements the DelegateAction[T] function required by
// Delegator[T]
type BaseIterator[T any] struct {
	// baseIterator is a pointed-to enclosed type due to BaseIterator providing
	// baseIterator.delegateAction
	*baseIterator[T]
}

// NewFunctionIterator returns a parli.Iterator based on a function.
// functionIterator is thread-safe and re-entrant.
func NewBaseIterator[T any](
	invokeFn InvokeFunc[T],
	delegateActionReceiver *DelegateAction[T],
) (iterator *BaseIterator[T]) {
	if invokeFn == nil {
		panic(perrors.NewPF("invokeFn cannot be nil"))
	} else if delegateActionReceiver == nil {
		panic(perrors.NewPF("delegateActionReceiver cannot be nil"))
	}

	// pointer to baseIterator.delegateAction so this method can
	// be provided
	//	- also causes werrWait to be pointed to, a used non-copy struct
	i := baseIterator[T]{invokeFn: invokeFn}
	i.errWait.Add(1)
	*delegateActionReceiver = i.delegateAction

	return &BaseIterator[T]{baseIterator: &i}
}

// Cond implements the condition statement of a Go “for” clause
//   - the iterationVariable is updated by being provided as a pointer.
//     iterationVariable cannot be nil
//   - errp is an optional error pointer receiving any errors during iterator execution
//   - condition is true if iterationVariable was assigned a value and the iteration should continue
func (i *baseIterator[T]) Cond(iterationVariablep *T, errp ...*error) (condition bool) {
	if iterationVariablep == nil {
		perrors.NewPF("iterationVariablep cannot bee nil")
	}

	// handle error
	if ep := i.err.Load(); ep != nil {
		if err := *ep; err != nil {
			if len(errp) > 0 {
				if errp0 := errp[0]; errp0 != nil {
					*errp0 = err // update errp with error
				}
			}
			return // error return: cond false, iterationVariablep unchanged, errp updated
		}
	}

	// check for next value
	var value T
	if value, condition = i.delegateAction(IsNext); condition {
		*iterationVariablep = value
	}

	return // condition and iterationVariablep updated, errp unchanged
}

// Cancel release resources for this iterator.
// Not every iterator requires a Cancel invocation.
func (i *baseIterator[T]) Cancel(errp ...*error) (err error) {

	// ignore if cancel alread invoked
	if !i.cancelState.CompareAndSwap(uint32(notCanceled), uint32(cancelRequested)) {
		i.errWait.Wait()
		if err = *i.err.Load(); err != nil {
			if len(errp) > 0 {
				if ep := errp[0]; ep != nil {
					*ep = perrors.AppendError(*ep, err)
				}
			}
		}
		return // already beyond cancel
	}

	// wait for access to fn
	i.publicsLock.Lock()
	defer i.publicsLock.Unlock()

	_, _, err = i.enqueueForFn(enqueueForFnCancel)
	if err != nil {
		if len(errp) > 0 {
			if ep := errp[0]; ep != nil {
				*ep = perrors.AppendError(*ep, err)
			}
		}
	}

	return
}

// delegateAction finds the next or the same value
//   - isSame true means first or same value should be returned
//   - value is the sought value or the T type’s zero-value if no value exists
//   - hasValue true means value was assigned a valid T value
func (i *baseIterator[T]) delegateAction(isSame NextAction) (value T, hasValue bool) {

	// fast outside-lock value-check
	if i.noValuesAvailable.Load() {
		return // no more values: zero-value and hasValue false
	}
	// wait for access to fn
	i.publicsLock.Lock()
	defer i.publicsLock.Unlock()

	// inside-lock check
	if i.noValuesAvailable.Load() {
		return // no more values: zero-value and hasValue false
	}
	if cancelStates(i.cancelState.Load()) != notCanceled {
		i.noValuesAvailable.Store(true)
		return // some cancellation of the iterator
	}

	// same value when a value has previously been read
	if i.didNext {
		if isSame == IsSame {
			value = i.value
			hasValue = i.hasValue
			return // same value return
		}
	} else {
		// did seek the first value
		i.didNext = true
	}

	// invoke fn
	var cancelState cancelStates
	var err error
	// enqueueForFn unlocks while invoking fn to allow re-entrancy.
	// if fn returns error, enqueueForFn updates: iter.isEnd iter.err iter.value iter.hasValue.
	value, cancelState, err = i.enqueueForFn(enqueueForFnNextItem)

	// determine whether a value was provided
	hasValue = err == nil && cancelState == notCanceled

	if !hasValue {
		i.noValuesAvailable.Store(true)
		return // no more values available
	}

	// store received value for future IsSame
	i.value = value
	i.hasValue = hasValue

	return // hasValue true, valid value return
}

// enqueueForFn invokes iter.fn
//   - is hasValue true, value is valid, err nil, cancelState notCanceled
//   - otherwise, i.err is updated, i.errWait released,
//     cancelState other than notCanceled or cancelRequested
//   - —
//   - invoked while holding i.publicsLock
//   - recovers from fn panics
//   - updates i.err i.cancelState i.errWait
func (i *baseIterator[T]) enqueueForFn(isCancel bool) (
	value T,
	cancelState cancelStates,
	err error,
) {

	// cancelState immediately prior to invoking i.invokeFn
	cancelState = cancelStates(i.cancelState.Load())

	// is invocation is still possible?
	if cancelState != notCanceled &&
		cancelState != cancelRequested {
		err = *i.err.Load() // collect previous error
		return              // no-further-invocations state return
	}
	// true if a panic occurred here or in invokeFn
	//	- false: invokeFn returned
	var isPanic bool
	// true if invokeFn decided to cancel istead of request a value
	var didCancel bool
	// deferred update of i.cancelState i.err i.errWait
	defer i.updateState(&cancelState, &didCancel, &err)
	// capture panics
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, &isPanic)

	// invoke i.invokeFn

	// wasCancel is the cancel value provided to invokeFn
	var wasCancel = cancelState != notCanceled
	// the value returned by invokeFn
	var v T

	v, didCancel, isPanic, err = i.invokeFn(wasCancel)

	// determine if value is valid
	if err == nil && !wasCancel && !didCancel {
		value = v
	}

	return
}

// update i.cancelState i.err i.errWait
func (i *baseIterator[T]) updateState(cancelStatep *cancelStates, didCancelp *bool, errp *error) {
	// cancelState immediately prior to invoking i.invokeFn
	//	- either notCanceled cancelRequested
	var cancelState = *cancelStatep
	// whether invokeFn decided to cancel iteration
	var didCancel = *didCancelp
	// possible error returned by i.invokeFn or result of panic
	var err = *errp

	plog.D("updateState: cancelState %s didCancel %t err %s",
		cancelState, didCancel, perrors.Short(err),
	)

	// determine next state and ignore ErrEndCallbacks
	var nextState = notCanceled
	if err != nil {

		// received a bad error from i.invokeFn
		if !errors.Is(err, cyclebreaker.ErrEndCallbacks) {
			nextState = errorReceived
		} else {

			// received end-of-data from i.invokeFn
			err = nil // ignore the error
			if cancelState == notCanceled {
				nextState = endOfData // spontaneous end of data
			} else {
				nextState = cancelComplete // non-error return from cancel invocation
			}
		}
	} else if cancelState == cancelRequested || didCancel {

		// non-error return from cancel invocation
		nextState = cancelComplete
	}

	// handle transition to ended state
	if nextState != notCanceled {
		// update the ended state
		i.cancelState.Store(uint32(nextState))
		*cancelStatep = nextState // update return value
		// try and store error
		if i.err.CompareAndSwap(nil, &err) {
			// if error update, notify threads awaiting error
			i.errWait.Done()
		}
	}
}
