/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"fmt"

	"github.com/haraldrudell/parl/perrors"
)

type SliceInterfaceIterator[I any, E any] struct {
	SlicePointerIterator[E]
	// Delegator implements the value methods required by the [Iterator] interface
	//   - Next HasNext NextValue
	//     Same Has SameValue
	//   - the delegate provides DelegateAction[T] function
	Delegator[I]
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
	i := SliceInterfaceIterator[I, E]{}
	// the delegator provides method based on the interface type
	i.Delegator = *NewDelegator[I](i.delegateAction)
	// the iterator returns *E values
	NewSlicePointerIteratorField(&i.SlicePointerIterator, slice)
	return &i
}

// Init initializes I interface values and returns an I iterator
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *SliceInterfaceIterator[I, E]) Init() (iterationVariable I, iterator Iterator[I]) {
	iterator = i
	return
}

// Cond updates I interface pointers
//
// Usage:
//
//		for i, iterator := NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
//	   // i is pointer to slice element
func (i *SliceInterfaceIterator[I, E]) Cond(iterationVariablep *I, errp ...*error) (condition bool) {
	var ep *E
	if condition = i.SlicePointerIterator.Cond(&ep); !condition {
		var iValue I
		*iterationVariablep = iValue // assign zero-value
		return                       // pointer assigned nil, condition: false
	}
	*iterationVariablep = any(ep).(I)
	return // pointer assigned asserted &E, condition: true
}

// delegateAction returns I values to the delegator
// by type asserting *E values from the slice pointer-iterator
func (i *SliceInterfaceIterator[I, E]) delegateAction(isSame NextAction) (value I, hasValue bool) {
	var ep *E
	if ep, hasValue = i.SlicePointerIterator.delegateAction(isSame); !hasValue {
		return
	}

	// here we want value to be *E
	//	- if ep is nil, asserting it will cause a typed nil
	//	- in Go, typed nil does not equal nil, so this must be avoided
	value = any(ep).(I) // check to not panic in new function

	return
}
