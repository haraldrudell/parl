/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// Function traverses a function generating values
type SimpleFunc[T any] struct {
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
	iteratorFunction SimpleIteratorFunc[T]
	// BaseIterator implements the DelegateAction[T] function required by
	// Delegator[T] and Cancel
	//	- provides its delegateAction method to Delegator
	*BaseIterator[T]
}

// NewFunctionIterator returns an [Iterator] iterating over a function
//   - thread-safe
func NewSimpleFunctionIterator[T any](
	iteratorFunction SimpleIteratorFunc[T],
) (iterator Iterator[T]) {
	if iteratorFunction == nil {
		panic(cyclebreaker.NilError("iteratorFunction"))
	}
	f := &SimpleFunc[T]{iteratorFunction: iteratorFunction}
	f.BaseIterator = NewBaseIterator[T](f.iteratorAction)
	return f
}

// Init implements the right-hand side of a short variable declaration in
// the init statement for a Go “for” clause
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *SimpleFunc[T]) Init() (iterationVariable T, iterator Iterator[T]) {
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
func (i *SimpleFunc[T]) iteratorAction(isCancel bool) (value T, isPanic bool, err error) {
	if isCancel {
		return
	}
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, &isPanic)

	// func() (value T, hasValue bool)
	var hasValue bool
	if value, hasValue = i.iteratorFunction(); !hasValue {
		err = cyclebreaker.ErrEndCallbacks
	}

	return
}
