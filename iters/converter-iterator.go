/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
)

// converterIterator is an enclosed implementation type for
// ConverterIterator[K, V]
type converterIterator[K constraints.Ordered, V any] struct {
	// keyIterator provides the key values converterFunction uses to
	// return values
	keyIterator Iterator[K]
	// ConverterFunction receives a key and returns the corresponding value.
	//   - if isCancel true, it means this is the last invocation of ConverterFunction and
	//     ConverterFunction should release any resources.
	//     Any returned value is not used
	//   - ConverterFunction signals end of values by returning parl.ErrEndCallbacks.
	//     if hasValue true, the accompanying value is used
	//   - if ConverterFunction returns error, it will not be invoked again.
	//     Any returned value is not used
	//   - ConverterFunction must be thread-safe
	//   - ConverterFunction is invoked by at most one thread at a time
	converterFunction ConverterFunction[K, V]
}

// ConverterIterator traverses another iterator and returns converted values. Thread-safe.
type ConverterIterator[K constraints.Ordered, V any] struct {
	// converterIterator implements the invokeConverterFunction method
	//	- pointer since invokeConverterFunction method is provided to
	//		BaseIterator
	*converterIterator[K, V]
	// BaseIterator implements Cancel and the DelegateAction[T] function required by
	// Delegator[T]
	//	- receives invokeConverterFunction function
	//	- provides delegateAction function
	BaseIterator[V]
	// Delegator implements the value methods required by the [Iterator] interface
	//   - Next HasNext NextValue
	//     Same Has SameValue
	//   - receives DelegateAction[T] function
	Delegator[V]
}

// NewConverterIterator returns a converting iterator.
//   - ConverterIterator is thread-safe and re-entrant.
func NewConverterIterator[K constraints.Ordered, V any](
	keyIterator Iterator[K],
	converterFunction ConverterFunction[K, V],
) (iterator Iterator[V]) {
	if converterFunction == nil {
		panic(perrors.NewPF("fn cannot be nil"))
	} else if keyIterator == nil {
		panic(perrors.NewPF("keyIterator cannot be nil"))
	}

	c := converterIterator[K, V]{
		keyIterator:       keyIterator,
		converterFunction: converterFunction,
	}

	var delegateAction DelegateAction[V]
	return &ConverterIterator[K, V]{
		converterIterator: &c,
		BaseIterator:      *NewBaseIterator(c.invokeConverterFunction, &delegateAction),
		Delegator:         *NewDelegator(delegateAction),
	}
}

// invokeConverterFunction invokes converterFunction recovering a possible panic
//   - if cancelState == notCanceled, a new value is requested.
//     Otherwise, iteration cancel is requested
//   - if err is nil, value is valid and isPanic false.
//     Otherwise, err is non-nil and isPanic may be set.
//     value is zero-value
func (i *converterIterator[K, T]) invokeConverterFunction(isCancel bool) (
	value T,
	didCancel bool,
	isPanic bool,
	err error,
) {
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, &isPanic)

	// get next key from keyIterator
	var key K
	if !isCancel {
		var hasKey bool
		key, hasKey = i.keyIterator.Next()
		isCancel = !hasKey
		didCancel = !hasKey
	}

	var v T
	// invoke converter function
	v, err = i.converterFunction(key, isCancel)

	// determine iof value is valid
	if err == nil && !isCancel {
		value = v
	}

	return
}
