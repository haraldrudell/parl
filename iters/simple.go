/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

// Converter traverses another iterator and returns converted values
type Simple[K any, V any] struct {
	// keyIterator provides the key values converterFunction uses to
	// return values
	keyIterator     Iterator[K]
	simpleConverter func(key K) (value V)
	// BaseIterator implements Cancel and the DelegateAction[T] function required by
	// Delegator[T]
	//	- receives invokeConverterFunction function
	//	- provides delegateAction function
	BaseIterator[V]
}

// NewConverterIterator returns a converting iterator.
//   - converterFunction receives cancel and can return error
//   - ConverterIterator is thread-safe and re-entrant.
//   - stores self-referencing pointers
func NewSimpleConverterIterator[K any, V any](
	keyIterator Iterator[K],
	simpleConverter SimpleConverter[K, V],
) (iterator Iterator[V]) {
	if simpleConverter == nil {
		panic(cyclebreaker.NilError("simpleConverter"))
	} else if keyIterator == nil {
		panic(cyclebreaker.NilError("keyIterator"))
	}

	c := Simple[K, V]{
		keyIterator:     keyIterator,
		simpleConverter: simpleConverter,
	}
	c.BaseIterator = *NewBaseIterator(c.iteratorAction)

	return &c
}

// Init implements the right-hand side of a short variable declaration in
// the init statement for a Go “for” clause
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *Simple[K, T]) Init() (iterationVariable T, iterator Iterator[T]) {
	iterator = i
	return
}

// iteratorAction invokes converterFunction recovering a possible panic
//   - if cancelState == notCanceled, a new value is requested.
//     Otherwise, iteration cancel is requested
//   - if err is nil, value is valid and isPanic false.
//     Otherwise, err is non-nil and isPanic may be set.
//     value is zero-value
func (i *Simple[K, T]) iteratorAction(isCancel bool) (value T, err error) {
	if isCancel {
		return
	}
	// get next key from keyIterator
	var key K
	var hasKey bool
	if key, hasKey = i.keyIterator.Next(); !hasKey {
		err = cyclebreaker.ErrEndCallbacks
		return
	}
	value = i.simpleConverter(key)

	return
}
