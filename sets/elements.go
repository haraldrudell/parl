/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
)

// Elements is an Iterator[Element[T]] that reads from concrete slice []E
//   - E is input concrete type, likely *struct, that implements Element[T]
//   - T is some comparable type that E produces, allowing for different
//     instances of E to be distinguished from one another
//   - Element[T] is a type used by sets, enumerations and bit-fields
//   - the Elements iterator provides a sequence of interface-type elements
//     based on a slice of concrete implementation-type values,
//     without re-creating the slice. Go cannot do type assertions on a slice,
//     only on individual values
type Elements[T comparable, E any] struct {
	// elementSlice implements Cancel and delegateAction function
	//	- implements iterator for Element[T]
	//	- pointer since delegateAction is provided to delegate
	*elementsAction[T, E]
	// Delegator implements the value methods required by the [Iterator] interface
	//   - Next HasNext NextValue
	//     Same Has SameValue
	//   - the delegate provides DelegateAction[T] function
	iters.Delegator[Element[T]]
}

// elementsAction is an enclosed type implementing delegateAction function
type elementsAction[T comparable, E any] struct {
	elements []E // the slice providing values

	// indicates that no further values can be returned
	//	- written behind publicsLock
	noValuesAvailable atomic.Bool

	// publicsLock serializes invocations of iterator [ElementSlice.delegateAction]
	publicsLock sync.Mutex

	// delegateAction

	// didNext indicates that a Next operation has completed and that hasValue may be valid
	//	- behind publicsLock
	didNext bool
	// index in slice, 0…len(slice)
	//	- behind lock
	index int
}

// NewElements returns an iterator of interface-type sets.Element[T]
//   - elements is a slice of a concrete type, named E, that should implement
//     sets.Element
//   - at compile time, elements is slice of any: []any
//   - based on a slice of non-interface-type Elements[T comparable].
func NewElements[T comparable, E any](elements []E) (iter iters.Iterator[Element[T]]) {

	// runtime check if required type conversion works
	var pointerToRuntimeE *E
	var pointerToRuntimeEInterface any = pointerToRuntimeE
	// type assertion of the runtime type *E to the interface type Element[T]
	if _, ok := pointerToRuntimeEInterface.(Element[T]); !ok {

		// runtime type *E does not implement Element[T]
		//	- produce error message

		// string description of runtime type Element[T]
		var runtimeTypeElementTStringDescription string
		var t *Element[T]
		if runtimeTypeElementTStringDescription = fmt.Sprintf("%T", t); len(runtimeTypeElementTStringDescription) > 0 {
			// drop leading * indicating pointer
			runtimeTypeElementTStringDescription = runtimeTypeElementTStringDescription[1:]
		}

		// string description of runtime type E
		var runtimeTypeEStringDescription string
		if runtimeTypeEStringDescription = fmt.Sprintf("%T", t); len(runtimeTypeEStringDescription) > 0 {
			// delete the * indicating pointer
			runtimeTypeEStringDescription = runtimeTypeEStringDescription[1:]
		}

		panic(perrors.ErrorfPF("input type %s does not implement interface-type %s",
			runtimeTypeEStringDescription,
			runtimeTypeElementTStringDescription,
		))
	}

	// create the iterator
	e := elementsAction[T, E]{elements: elements}

	return &Elements[T, E]{
		elementsAction: &e,
		Delegator:      *iters.NewDelegator(e.delegateAction),
	}
}

// delegateAction finds the next or the same value. Thread-safe
//   - isSame == IsSame means first or same value should be returned
//   - value is the sought value or the T type’s zero-value if no value exists
//   - hasValue true means value was assigned a valid T value
func (i *elementsAction[T, E]) delegateAction(isSame iters.NextAction) (value Element[T], hasValue bool) {

	// fast outside-lock value-check
	if i.noValuesAvailable.Load() {
		return // no more values return
	}

	i.publicsLock.Lock()
	defer i.publicsLock.Unlock()

	// inside-lock value-check
	if i.noValuesAvailable.Load() {
		return // no more values return
	}

	// for IsSame operation the first value must be sought
	//	- therefore, if the first value has not been sought, seek it now or
	//	- if not IsSame operation, advance to the next value
	if i.didNext {
		if isSame == iters.IsNext {
			// find slice index to use
			//	- advance index until final value len(i.slice)
			if i.index < len(i.elements) {
				i.index++
			}
		}
	} else {
		// note that first value has been sought
		i.didNext = true
	}

	// check if the new index is within available slice values
	if hasValue = i.index < len(i.elements); !hasValue {
		i.noValuesAvailable.CompareAndSwap(false, true)
		return // no values return
	}

	// a is any but runtime type is *E, pointer to a concrete type, likely *struct
	var ePointer any = &i.elements[i.index]

	// do type assertion of ePointer to interface Element[T]
	var ok bool
	if value, ok = ePointer.(Element[T]); !ok {
		// this type assertion was checked by NewElements: should never happen
		panic(perrors.ErrorfPF("type assertion failed: %T %T", ePointer, value))
	}

	return // hasValue true, value valid return
}

// Cancel release resources for this iterator. Thread-safe
//   - not every iterator requires a Cancel invocation
func (i *elementsAction[T, E]) Cancel() (err error) {
	i.noValuesAvailable.CompareAndSwap(false, true)
	return
}
