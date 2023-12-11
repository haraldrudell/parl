/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

// Converter traverses another iterator and returns converted values
type Converter[K any, V any] struct {
	// keyIterator provides the key values converterFunction uses to
	// return values
	keyIterator Iterator[K]
	// ConverterFunction receives a key and returns the corresponding value
	//	- func(key K, isCancel bool) (value V, err error)
	//   - if isCancel true, it means this is the last invocation of ConverterFunction and
	//     ConverterFunction should release any resources.
	//     Any returned value is not used
	//   - ConverterFunction signals end of values by returning parl.ErrEndCallbacks.
	//     if hasValue is true, the accompanying value is used
	//   - when ConverterFunction returns error, it will not be invoked again.
	//     For errors other than parl.ErrEndCallbacks, value is not used
	//   - ConverterFunction must be thread-safe
	//   - ConverterFunction is invoked by at most one thread at a time
	converter ConverterFunction[K, V]
	*BaseIterator[V]
}

// NewConverterIterator returns a converting iterator.
//   - converterFunction receives cancel and can return error
//   - ConverterIterator is thread-safe and re-entrant.
//   - stores self-referencing pointers
func NewConverterIterator[K any, V any](
	keyIterator Iterator[K],
	converter ConverterFunction[K, V],
) (iterator Iterator[V]) {
	if converter == nil {
		panic(cyclebreaker.NilError("converter"))
	} else if keyIterator == nil {
		panic(cyclebreaker.NilError("keyIterator"))
	}

	c := Converter[K, V]{
		keyIterator: keyIterator,
		converter:   converter,
	}
	c.BaseIterator = NewBaseIterator(c.iteratorAction)

	return &c
}

// Init implements the right-hand side of a short variable declaration in
// the init statement for a Go “for” clause
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *Converter[K, T]) Init() (iterationVariable T, iterator Iterator[T]) {
	iterator = i
	return
}

// iteratorAction invokes converterFunction recovering a possible panic
//   - if cancelState == notCanceled, a new value is requested.
//     Otherwise, iteration cancel is requested
//   - if err is nil, value is valid and isPanic false.
//     Otherwise, err is non-nil and isPanic may be set.
//     value is zero-value
func (i *Converter[K, T]) iteratorAction(isCancel bool) (value T, err error) {

	// get next key from keyIterator
	var key K
	var isEndOfKeys bool
	if !isCancel {
		var hasKey bool
		if key, hasKey = i.keyIterator.Next(); !hasKey {
			isEndOfKeys = true
			isCancel = true
		}
	}

	// invoke converter function
	//	- func(key K, isCancel bool) (value V, err error)
	if value, err = i.converter(key, isCancel); err != nil {
		err = perrors.Stack(err)
	} else if isEndOfKeys {
		err = cyclebreaker.ErrEndCallbacks
	}

	return
}
