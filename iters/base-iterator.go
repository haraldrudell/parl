/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"errors"
	"sync"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// BaseIterator is a type enclosed in iterator implementations implementing:
//   - [Iterator] methods [Iterator.Cond] [Iterator.Next] [Iterator.Cancel]
//   - consumer must:
//   - — implement [Iterator.Init] since it returns the enclosing type
//   - — provide [IteratorAction]
type BaseIterator[T any] struct {
	// iteratorAction is value-retrieving implementation provided by the consumer
	//	- on invocation, iteratorAction returns the next value or error
	//	- iteratorAction is critial section
	iteratorAction IteratorAction[T]

	// cancelState controls the behavior of the iterator
	//	- initial cancelState is notCanceled meaning iteration is in progress
	//	- [BaseIterator.Cancel] when notCanceled goes ot state cancelRequested outside any lock.
	//		Cancel also invokes asyncCancel if present, every time
	//	- doAction provides the cancel request to iteratorAction behind publicsLock
	//	- — if iteratorAction returns error, state is errorReceived
	//	- — if iteratorAction returns without error from cancelRequested, state is cancelComplete
	//	- — if iteratorAction returns ErrEndCallbacks, state is endOfData
	//	- the end state provide information as to how iteration ended
	//	- written behind publicsLock except Cancel CompareAndSwap
	//	- not publicly accessible
	cancelState cyclebreaker.Atomic32[cancelStates]
	// publicsLock serializes invocations of [BaseIterator.Next] and [BaseIterator.Cancel]
	publicsLock sync.Mutex
	// err is a thread-safe error container
	//	- recover value on panic, cancelState is panicked
	//	- any returned error from iteratorAction other than [parl.ErrEndCallbacks],
	//		cancelState is errorReceived
	//	- if non-nil, further [BaseIterator.Next] are prevented
	//	- read outside lock by [BaseIterator.Next] [BaseIterator.Cond] [BaseIterator.Cancel]
	err cyclebreaker.AtomicError
	// asyncCancel is invoked on cancel if present, allowing a blocking iterator
	// to abort an operation i progress
	asyncCancel func()
}

// NewBaseIterator returns an implementation of Cond Next Cancel methods part of [iters.Iterator]
//   - asyncCancel is used if function or converter iterators are blocking.
//     asyncCancel indicates that a Cancel invocation has occurred.
//     ayncCancel may close a blocking Reader thereby asynchorously ending iteration
func NewBaseIterator[T any](
	iteratorAction IteratorAction[T],
	asyncCancel ...func(),
) (iterator *BaseIterator[T]) {
	if iteratorAction == nil {
		panic(cyclebreaker.NilError("iteratorAction"))
	}
	var ac func()
	if len(asyncCancel) > 0 {
		ac = asyncCancel[0]
	}
	return &BaseIterator[T]{
		iteratorAction: iteratorAction,
		asyncCancel:    ac,
	}
}

// Cond implements the condition statement of a Go “for” clause
//   - condition is true if iterationVariable was assigned a value and the iteration should continue
//   - the iterationVariable is updated by being provided as a pointer.
//     iterationVariable is only updated when condition is true.
//     iterationVariable cannot be nil
//   - errp is an optional error pointer receiving any errors during iterator execution.
//     errp is only updated when condition is false and an error has occurred
//
// Usage:
//
//	for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	  // i is pointer to slice element
func (i *BaseIterator[T]) Cond(iterationVariablep *T, errp ...*error) (condition bool) {
	if iterationVariablep == nil {
		cyclebreaker.NilError("iterationVariablep")
	}

	// next value
	var value T
	if value, condition = i.Next(); condition {
		*iterationVariablep = value
		return // condition true return: iterationVariablep updated
	}

	// no request to collect error
	if len(errp) == 0 {
		return // no more values return
	}

	// collect any error for errp
	i.getErr(replaceErrp, errp...)

	return // end of iteration, possible error collection return
}

// Next advances to next item and returns it
//   - if hasValue true, value contains the next value
//   - otherwise, no more items exist and value is the data type zero-value
func (i *BaseIterator[T]) Next() (value T, hasValue bool) {

	// fast outside-lock value-check
	if i.cancelState.Load() != notCanceled {
		return // no more values: zero-value and hasValue false
	}
	i.publicsLock.Lock()
	defer i.publicsLock.Unlock()

	// inside-lock check
	if i.cancelState.Load() != notCanceled {
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
	if i.cancelState.Load() == notCanceled {
		i.cancelState.CompareAndSwap(notCanceled, cancelRequested)
	}
	// asyncCancel
	if ac := i.asyncCancel; ac != nil {
		ac()
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
	var err error
	var isPanic bool
	defer i.doActionEnd(&isPanic, &err)
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, &isPanic)

	// cancelState inside lock prior to invoking iteratorAction
	var cancelState = i.cancelState.Load()
	if cancelState == cancelRequested {
		action = doCancel // after Cancel invoked, the only action is cancel
	}

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
		// ensure correct state prior to updating error
		i.cancelState.Store(nextState)
		i.err.AddError(err)
	}

	return
}

func (i *BaseIterator[T]) doActionEnd(isPanic *bool, errp *error) {
	if !*isPanic {
		return
	}
	i.err.AddError(*errp)
	i.cancelState.Store(panicked)
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
//   - errp: optional error pointer. if present and non-nil, a possible error is appended here
//   - replaceErrp0 true: when updating errp, the error at *errp is replaced, no appended to
//   - err: a possible error retieved from the error store
//   - isPresent:
func (i *BaseIterator[T]) getErr(replaceErrp0 getErrAppend, errp ...*error) (isPresent bool, err error) {
	if err, isPresent = i.err.Error(); !isPresent {
		return // err is not present yet
	}

	// save error to *errp if possible
	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil {
			// errp is present and non-nil
			if replaceErrp0 {
				*errp0 = err // update errp with error
			} else {
				*errp0 = perrors.AppendError(*errp0, err)
			}
		}
	}

	return // error return
}
