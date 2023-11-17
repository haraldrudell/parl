/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// ConverterIterator traverses another iterator and returns converted values
type ConverterIterator[K any, V any] struct {
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
	simpleConverter   func(key K) (value V)
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
//   - converterFunction receives cancel and can return error
//   - ConverterIterator is thread-safe and re-entrant.
//   - stores self-referencing pointers
func NewConverterIterator[K any, V any](
	keyIterator Iterator[K],
	converterFunction ConverterFunction[K, V],
) (iterator Iterator[V]) {
	if converterFunction == nil {
		panic(perrors.NewPF("fn cannot be nil"))
	} else if keyIterator == nil {
		panic(perrors.NewPF("keyIterator cannot be nil"))
	}

	c := ConverterIterator[K, V]{
		keyIterator:       keyIterator,
		converterFunction: converterFunction,
	}
	var delegateAction DelegateAction[V]
	c.BaseIterator = *NewBaseIterator(c.invokeConverterFunction, &delegateAction)
	c.Delegator = *NewDelegator(delegateAction)

	return &c
}

// NewConverterIteratorField returns a converting iterator
//   - if fieldp is non-nil, this field is initialized
//   - either converterFunction or simpleConverter must be non-nil
//   - if simpleConverter is non-nil it is used.
//     simpleConverter is simplified in that it cannot return error and has no concept of cancel.
//     simpleConverter must on each invocation return the value corresponding to the key provided
//   - converterFunction receives cancel and can return error
//   - ConverterIterator is thread-safe and re-entrant.
//   - stores self-referencing pointers
func NewConverterIteratorField[K any, V any](
	fieldp *ConverterIterator[K, V],
	keyIterator Iterator[K],
	converterFunction ConverterFunction[K, V],
	simpleConverter func(key K) (value V),
) (iterator Iterator[V]) {
	if converterFunction == nil && simpleConverter == nil {
		panic(perrors.NewPF("converterFunction and simpleConverter cannot both be nil"))
	} else if keyIterator == nil {
		panic(perrors.NewPF("keyIterator cannot be nil"))
	}
	var c *ConverterIterator[K, V]
	if fieldp != nil {
		c = fieldp
	} else {
		c = &ConverterIterator[K, V]{}
	}
	iterator = c
	c.keyIterator = keyIterator
	c.converterFunction = converterFunction
	c.simpleConverter = simpleConverter
	var delegateAction DelegateAction[V]
	c.BaseIterator = *NewBaseIterator(c.invokeConverterFunction, &delegateAction)
	c.Delegator = *NewDelegator(delegateAction)
	return
}

// Init implements the right-hand side of a short variable declaration in
// the init statement for a Go “for” clause
func (i *ConverterIterator[K, T]) Init() (iterationVariable T, iterator Iterator[T]) {
	iterator = i
	return
}

// invokeConverterFunction invokes converterFunction recovering a possible panic
//   - if cancelState == notCanceled, a new value is requested.
//     Otherwise, iteration cancel is requested
//   - if err is nil, value is valid and isPanic false.
//     Otherwise, err is non-nil and isPanic may be set.
//     value is zero-value
func (i *ConverterIterator[K, T]) invokeConverterFunction(isCancel bool) (
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

	// simple converter case: no cancel, value always returned
	if simpleConverter := i.simpleConverter; simpleConverter != nil {
		if !isCancel {
			value = simpleConverter(key)
		}
		return
	}

	var v T
	// invoke converter function
	v, err = i.converterFunction(key, isCancel)

	// determine if value is valid
	if err == nil && !isCancel {
		value = v
	}

	return
}
