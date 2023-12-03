/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// Function traverses a function generating values
type Function[T any] struct {
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
	// BaseIterator implements the DelegateAction[T] function required by
	// Delegator[T] and Cancel
	//	- provides its delegateAction method to Delegator
	*BaseIterator[T]
}

// NewFunctionIterator returns an [Iterator] iterating over a function
//   - thread-safe
func NewFunctionIterator[T any](
	iteratorFunction IteratorFunction[T],
) (iterator Iterator[T]) {
	if iteratorFunction == nil {
		panic(cyclebreaker.NilError("iteratorFunction"))
	}
	f := &Function[T]{iteratorFunction: iteratorFunction}
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
func (i *Function[T]) Init() (iterationVariable T, iterator Iterator[T]) {
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
func (i *Function[T]) iteratorAction(isCancel bool) (value T, isPanic bool, err error) {
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, &isPanic)

	// func(isCancel bool) (value T, err error)
	if value, err = i.iteratorFunction(isCancel); err != nil {
		err = perrors.Stack(err)
	}

	return
}
