/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"errors"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// functionIterator is an enclosed implementation type for
// FunctionIterator[T]
type functionIterator[T any] struct {
	// IteratorFunction is a function that can be used with function iterator
	//   - if isCancel true, it means this is the last invocation of IteratorFunction and
	//     IteratorFunction should release any resources.
	//     Any returned value is not used
	//   - IteratorFunction signals end of values by returning parl.ErrEndCallbacks.
	//     if hasValue true, the accompanying value is used
	//   - if IteratorFunction returns error, it will not be invoked again.
	//     Any returned value is not used
	//   - IteratorFunction must be thread-safe
	//   - IteratorFunction is invoked by at most one thread at a time
	iteratorFunction IteratorFunction[T]
}

// FunctionIterator traverses a function generating values
type FunctionIterator[T any] struct {
	// functionIterator invokes IteratorFunction[T]
	//	- pointer since it provides its invokeFn method to BaseIterator
	*functionIterator[T]
	// BaseIterator implements the DelegateAction[T] function required by
	// Delegator[T] and Cancel
	//	- provides its delegateAction method to Delegator
	BaseIterator[T]
	// Delegator implements the value methods required by the [Iterator] interface
	//   - Next HasNext NextValue
	//     Same Has SameValue
	//   - Delegator obtains items using the provided DelegateAction[T] function
	Delegator[T]
}

// NewFunctionIterator returns an [Iterator] iterating over a function
//   - thread-safe
func NewFunctionIterator[T any](
	iteratorFunction IteratorFunction[T],
) (iterator Iterator[T]) {
	if iteratorFunction == nil {
		panic(perrors.NewPF("fn cannot be nil"))
	}

	f := functionIterator[T]{iteratorFunction: iteratorFunction}

	var delegateAction DelegateAction[T]
	return &FunctionIterator[T]{
		functionIterator: &f,
		BaseIterator:     *NewBaseIterator[T](f.baseIteratorRequest, &delegateAction),
		// NewDelegator must be after NewBaseIterator
		Delegator: *NewDelegator[T](delegateAction),
	}
}

// Init implements the right-hand side of a short variable declaration in
// the init statement for a Go “for” clause
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *FunctionIterator[T]) Init() (iterationVariable T, iterator Iterator[T]) {
	iterator = i
	return
}

// baseIteratorRequest invokes fn recovering a possible panic
//   - if cancelState == notCanceled, a new value is requested.
//     Otherwise, iteration cancel is requested
//   - if err is nil, value is valid and isPanic false.
//     Otherwise, err is non-nil and isPanic may be set.
//     value is zero-value
//   - thread-safe but invocations must be serialized
func (i *functionIterator[T]) baseIteratorRequest(isCancel bool) (
	value T,
	didCancel bool,
	isPanic bool,
	err error,
) {
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, &isPanic)

	value, err = i.iteratorFunction(isCancel)

	// baseIterator will filter value based on validity

	// iteratorFunction may return ErrEndCallback
	//	- other delegates of baseIterator does not do this,
	//		has to be handled here
	if err != nil {
		if errors.Is(err, cyclebreaker.ErrEndCallbacks) {
			didCancel = true // iterator function did cancel
			err = nil        // ignore the error
		} else {
			err = perrors.Stack(err)
		}
	}

	return
}
