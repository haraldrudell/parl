/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"fmt"

	"github.com/haraldrudell/parl/perrors"
)

type SliceInterface[I any, E any] struct {
	SlicePointer[E]
}

// NewSliceInterfaceIterator returns an iterator over slice []E returning those elements
// as interface I values
//   - [I ~*E, E any] as an expression of interface I implemented by &E
//     does not actually hold true at compile time.
//     The check whether I is an interface implemented by &E
//     has to take place at runtime
//   - uses self-referencing pointers
func NewSliceInterfaceIterator[I any, E any](slice []E) (iterator Iterator[I]) {

	// check (&E).(I) type assertion
	//	- I should be an interface type
	//	- E should be a concrete type implementing the I interface
	var ep *E
	var ok bool
	if _, ok = any(ep).(I); !ok {
		var ip *I
		var iType = fmt.Sprintf("%T", ip)[1:] // get the interface type name by removing leading star
		var e E
		var eType = fmt.Sprintf("%T", e)
		panic(perrors.ErrorfPF("I type: %s is not an interface that &E implements. E type: %s",
			iType, eType,
		))
	}

	// wrap a slice-pointer iterator in a type assertion shim
	i := SliceInterface[I, E]{}
	// the iterator returns *E values
	NewSlicePointerIteratorField(&i.SlicePointer, slice)
	return &i
}

// Init initializes I interface values and returns an I iterator
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *SliceInterface[I, E]) Init() (iterationVariable I, iterator Iterator[I]) {
	iterator = i
	return
}

// Cond updates I interface pointers
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *SliceInterface[I, E]) Cond(iterationVariablep *I, errp ...*error) (condition bool) {
	var ep *E // getting *E but must be type asserted for Go to understand it’s interface I
	if condition = i.SlicePointer.Cond(&ep, errp...); !condition || ep == nil {
		var iZeroValue I
		*iterationVariablep = iZeroValue // assign zero-value
		return                           // pointer assigned nil, condition: valid
	}
	// ep is not nil, so assertion does not create typed nil
	*iterationVariablep = any(ep).(I)
	return // pointer assigned asserted &E, condition: true
}

// Next advances to next item and returns it
//   - if hasValue true, value contains the next value
//   - otherwise, no more items exist and value is the data type zero-value
func (i *SliceInterface[I, E]) Next() (value I, hasValue bool) { return i.nextSame(IsNext) }

// Same returns the same value again
//   - if hasValue true, value is valid
//   - otherwise, no more items exist and value is the data type zero-value
//   - If Next or Cond has not been invoked, Same first advances to the first item
func (i *SliceInterface[I, E]) Same() (value I, hasValue bool) { return i.nextSame(IsSame) }

// delegateAction returns I values to the delegator
// by type asserting *E values from the slice pointer-iterator
func (i *SliceInterface[I, E]) nextSame(isSame NextAction) (value I, hasValue bool) {
	var ep *E // getting *E but must be type asserted for Go to understand it’s interface I
	if ep, hasValue = i.SlicePointer.nextSame(isSame); !hasValue || ep == nil {
		return
	}

	// here we want value to be *E
	//	- if ep is nil, asserting it will cause a typed nil
	//	- in Go, typed nil does not equal nil, so this must be avoided
	value = any(ep).(I) // type assertion was checked to not panic in new function

	return
}
