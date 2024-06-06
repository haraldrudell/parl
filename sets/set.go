/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"fmt"

	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfmt"
	"github.com/haraldrudell/parl/plog"
)

// Set0 holds a multiple-selection function argument that has printable representation.
// PrintableEnum allows finite selection function arguments to have meaningful names and makes
// those names printable.
type Set0[T comparable, E any] struct {
	elementMap map[T]*E
	elements   []E
}

// NewSet returns an enumeration of printable semantic elements
//   - T is a comparable type for indexing the set, ie.
//     the type of the constants used as set values [SetElement.ValueV]
//   - E is type of declarative elements assigning additional properties to T.
//     Typically: [sets.SetElement].
//     E is typically inferred and does not have to pbe provided.
//   - elements: a slice whose elements implement the sets.Element[E] interface:
//   - — [SetElement.Value] method returns the T comparable
//   - — [SetElement.String] method returns string representation
//   - elements: a simple concrete type is [sets.SetElement]:
//   - — ValueV field is T comparable
//   - — Name field is the string value: typically single word
//   - another concerete type is sets.ElementFull[E]: Value, String, Description
//   - — Description method returns a sentence-length description
//
// usage:
//
//	sets.NewSet([]sets.SetElement[int]{
//		{ValueV: unix.AF_INET, Name: "IPv4"},
//		{ValueV: unix.AF_INET6, Name: "IPv6"},
//	})
func NewSet[T comparable, E any](elements []E) (set Set[T]) {
	return NewSetFieldp[T](elements, nil, nil, nil)
}

func NewSetFieldp[T comparable, E any](
	elements []E,
	fieldp *Set0[T, E],
	fTEp *func(comparable T) (ep *E),
	epIteratorp *func() (iterator iters.Iterator[*E]),
) (set Set[T]) {

	// E can be concrete type or interface but must have:
	//	- a Value method returning comparable T
	//	- a String method returning string
	var ep *E
	if _, ok := any(ep).(Element[T]); !ok {
		var eType = fmt.Sprintf("%T", ep)[1:] // get the interface type name by removing leading star
		var i Element[T]
		var interfaceType = fmt.Sprintf("%T", i)
		panic(perrors.ErrorfPF("type E %s: &E does not implement interface %s",
			eType, interfaceType,
		))
	}
	var m = make(map[T]*E, len(elements))
	for i := 0; i < len(elements); i++ {
		var ep = &elements[i]
		var element = any(ep).(Element[T])
		var valueT = element.Value()
		if existingElement, ok := m[valueT]; ok {
			panic(perrors.ErrorfPF(
				"duplicate set-element:"+
					" comparable type T: %T duplicate value: ‘%s’"+
					" new duplicate name: %q existing name: %q"+
					" number of added values: %d",
				valueT, pfmt.NoRecurseVPrint(valueT),
				element, existingElement,
				len(m),
			))
		}
		m[valueT] = ep
	}
	if fieldp != nil {
		*fieldp = Set0[T, E]{
			elementMap: m,
			elements:   elements,
		}
	} else {
		fieldp = &Set0[T, E]{
			elementMap: m,
			elements:   elements,
		}
	}
	if fTEp != nil {
		*fTEp = fieldp.fTE
	}
	if epIteratorp != nil {
		*epIteratorp = fieldp.epIterator
	}
	return fieldp
}

// IsValid returns whether value is part of the set
func (s *Set0[T, E]) IsValid(value T) (isValid bool) {
	_, isValid = s.elementMap[value]
	return
}

// Iterator allows iteration over all set elements
func (s *Set0[T, E]) Iterator() (iterator iters.Iterator[T]) {
	return iters.NewSimpleConverterIterator(
		iters.NewSlicePointerIterator(s.elements),
		s.simpleConverter,
	)
}

func (s *Set0[T, E]) simpleConverter(element *E) (valueT T) { return any(element).(Element[T]).Value() }

// StringT returns a string representation for an element of this set.
// If value is not a valid element, a fmt.Printf value is output like ?'%v'
func (s *Set0[T, E]) StringT(value T) (s2 string) {
	// StringT is intended to be the String method of a named type implementing set.
	// if StringT method code would somehow invoke the T.String method again,
	// this will cause infinite recursion and stack overflow panic.
	var ep *E
	var ok bool
	if ep, ok = s.elementMap[value]; ok {
		var element = any(ep).(Element[T])
		// e is Element interface, likely set.Element runtime type.
		// set.Element.String() returns a value from a type string, ie. type T
		// is not involved.
		// This ensures no T-type String-function recursion.
		s2 = element.String()
	} else {
		// converting T to string may recurse if printf %v or %s verbs are used.
		// ie. printf will re-enter this method and exhaust the stack.
		// all that is known about T is that it is a comparable type
		//p = string(value)

		s2 = "?\x27" + pfmt.NoRecurseVPrint(value) + "\x27"
	}
	return
}

// Description returns a more elaborate string representation
// for an element. Description and StringT may return the same value
func (s *Set0[T, E]) Description(value T) (full string) {

	// get a pointer to the Element type
	var ep *E
	var ok bool
	if ep, ok = s.elementMap[value]; !ok {
		return // invalid T return
	}

	// type assert to Element with Description method
	var elementDescription ElementDescription
	if elementDescription, ok = any(ep).(ElementDescription); !ok {
		return // not a full element return
	}

	full = elementDescription.Description()

	return
}

func (s *Set0[T, E]) fTE(valueT T) (ep *E) { return s.elementMap[valueT] }

func (s *Set0[T, E]) epIterator() (iterator iters.Iterator[*E]) {
	return iters.NewSlicePointerIterator(s.elements)
}

func (s *Set0[T, E]) String() (s2 string) {
	var t T
	s2 = plog.Sprintf("set_%T:%d", t, len(s.elementMap))
	return
}
