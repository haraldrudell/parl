/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// BaseIterator implements:
//   - [Iterator] methods [Iterator.Cond] [Iterator.Next] [Iterator.Cancel]
//   - consumer must:
//   - — implement [Iterator.Init] since it returns the enclosing type
//   - — provide IteratorAction[T]
type BaseIterator[T any] struct {
	iteratorAction IteratorAction[T]

	// cancelState is updated by:
	//	- Cancel() and
	//	- enqueueForFn() inside i.publicsLock lock
	//	- is state because Cancel is not complete until iteratorAction has returned
	cancelState atomic.Uint32
	// publicsLock serializes invocations of i.Next and i.Cancel
	publicsLock sync.Mutex
	// err contains any errors returned by fn other than parl.ErrEndCallbacks
	//	- nil if fn is active
	//	- non-nil pointer to nil error if fn completed without error
	//	- — returned parl.ErrEndCallbacks
	//	- — returned from FunctionIteratorCancel -1 invocation without error
	//	- non-nil error if fn returned error
	//	- updated by enqueueForFn() inside i.publicsLock lock
	err atomic.Pointer[error]
}

// NewBaseIterator returns an implementation of Cond Next Cancel methods part of [iters.Iterator]
func NewBaseIterator[T any](iteratorAction IteratorAction[T]) (iterator *BaseIterator[T]) {
	if iteratorAction == nil {
		panic(cyclebreaker.NilError("iteratorAction"))
	}
	return &BaseIterator[T]{iteratorAction: iteratorAction}
}

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
func (i *BaseIterator[T]) Cond(iterationVariablep *T, errp ...*error) (condition bool) {
	if iterationVariablep == nil {
		cyclebreaker.NilError("iterationVariablep")
	}

	// outside lock check updating errp
	if hasError, _ := i.getErr(replaceErrp, errp...); hasError {
		return // iterator is canceled
	}

	// next value
	var value T
	if value, condition = i.Next(); condition {
		*iterationVariablep = value
	} else if len(errp) > 0 {
		// collect any error for errp
		i.getErr(replaceErrp, errp...)
	}

	return // condition and iterationVariablep updated, errp unchanged
}

// Next advances to next item and returns it
//   - if hasValue true, value contains the next value
//   - otherwise, no more items exist and value is the data type zero-value
func (i *BaseIterator[T]) Next() (value T, hasValue bool) {

	// fast outside-lock value-check
	if i.err.Load() != nil {
		return // no more values: zero-value and hasValue false
	}
	i.publicsLock.Lock()
	defer i.publicsLock.Unlock()

	// inside-lock check
	if i.err.Load() != nil {
		return // no more values: zero-value and hasValue false
	}
	value, hasValue = i.doAction(doNext)

	return // hasValue true, valid value return
}

// Cancel stops an iteration
//   - after Cancel invocation, Cond, Next and Same indicate no value available
//   - Cancel returns the first error that occurred during iteration, if any
//   - an iterator implementation may require Cancel invocation
//     to release resources
//   - Cancel is deferrable
func (i *BaseIterator[T]) Cancel(errp ...*error) (err error) {

	// fast outside-lock check
	var hasErr bool
	if hasErr, err = i.getErr(appendErrp, errp...); hasErr {
		return // already canceled
	}
	// ensure cancel initiated prior to lock
	if cancelStates(i.cancelState.Load()) == notCanceled {
		i.cancelState.CompareAndSwap(uint32(notCanceled), uint32(cancelRequested))
	}
	// the lock provides wait mechanic
	i.publicsLock.Lock()
	defer i.publicsLock.Unlock()

	// inside lock check
	if hasErr, err = i.getErr(appendErrp, errp...); hasErr {
		return // already canceled
	}
	// send cancel to iteratorAction
	i.doAction(doCancel)
	_, err = i.getErr(appendErrp, errp...)

	return
}

const (
	// [BaseIterator.doAction] invoked for next value
	doNext anAction = false
	// [BaseIterator.doAction] invoked for cancel
	doCancel anAction = true
)

// Action for [BaseIterator.enqueueForFn]
//   - doNext doCancel
type anAction bool

// doAction invokes [BaseIterator.iteratorAction]
//   - invoked while holding [BaseIterator.publicsLock]
//   - [BaseIterator.err] should have been checked inside lock
//   - on return:
//   - — if hasValue is true, value is valid
//   - — [BaseIterator.err] may be updated
//   - —
//   - recovers from [BaseIterator.iteratorAction] panic
func (i *BaseIterator[T]) doAction(action anAction) (value T, hasValue bool) {

	// cancelState inside lock prior to invoking iteratorAction
	var cancelState = cancelStates(i.cancelState.Load())
	if cancelState == cancelRequested {
		action = doCancel // after Cancel invoked, the only action is cancel
	}
	defer cyclebreaker.Recover(func() cyclebreaker.DA { return cyclebreaker.A() }, nil, i.setErr)

	v, err := i.iteratorAction(bool(action))

	// determine if value is valid
	if hasValue = action == doNext && err == nil; hasValue {
		value = v
	}

	// update state: cancelComplete endOfData errorReceived
	var nextState cancelStates
	if err != nil {
		if !errors.Is(err, cyclebreaker.ErrEndCallbacks) {
			// received an unknown error
			nextState = errorReceived
		} else {
			// received end-of-data from iteratorAction
			err = nil // ignore the error
			if cancelState == notCanceled {
				// spontaneous end of data
				nextState = endOfData
			} else {
				// end-of-data return from cancel invocation
				nextState = cancelComplete
			}
		}
	} else if cancelState == cancelRequested {
		// non-error return from cancel invocation
		nextState = cancelComplete
	}
	if nextState != notCanceled {
		i.setErr2(err, nextState)
	}

	return
}

const (
	// [BaseIterator.getErr] uses [perrors.AppendError]
	appendErrp getErrAppend = false
	// [BaseIterator.getErr] replaces *errp
	replaceErrp getErrAppend = true
)

// how [BaseIterator.getErr] updates *errp
//   - appendErrp replaceErrp
type getErrAppend bool

// getErr reads [BaseIterator.err] and returns err, *errp
func (i *BaseIterator[T]) getErr(replace getErrAppend, errp ...*error) (isPresent bool, err error) {
	var ep = i.err.Load()
	if isPresent = ep != nil; !isPresent {
		return // err is not present yet
	} else if err = *ep; err == nil {
		return // result is no error
	}
	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil {
			if replace {
				*errp0 = err // update errp with error
			} else {
				*errp0 = perrors.AppendError(*errp0, err)
			}
		}
	}
	return // error return
}

// setErr updates atomic error container
func (i *BaseIterator[T]) setErr(err error) { i.setErr2(err, panicked) }

// setErr2 updates atomic error container
func (i *BaseIterator[T]) setErr2(err error, state cancelStates) {
	i.cancelState.Store(uint32(state))
	i.err.Store(&err)
}
