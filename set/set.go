/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package set

import (
	"strings"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfmt"
)

// SetImpl holds a multiple-selection function argument that has printable representation.
// PrintableEnum allows finite selection function arguments to have meaningful names and makes
// those names printable.
type SetImpl[T comparable] struct {
	elementMap map[T]Element[T]
}

// NewSet returns an enumeration of a printable semantic function argument.
// elements are the elements that form this set.
func NewSet[T comparable](elements []Element[T]) (interfaceSet Set[T]) {
	length := len(elements)
	set := SetImpl[T]{elementMap: map[T]Element[T]{}}
	for i := 0; i < length; i++ {
		ep := elements[i]
		valueT := ep.Value()
		if existingElement, ok := set.elementMap[valueT]; ok {
			panic(perrors.ErrorfPF(
				"duplicate set-element: type T: %T provided value: '%s' "+
					"provided name: %q existing name: %q "+
					"number of added values: %d",
				valueT, pfmt.NoRecurseVPrint(valueT), ep,
				existingElement,
				len(set.elementMap),
			))
		}
		set.elementMap[valueT] = ep
	}
	interfaceSet = &set
	return
}

func (st *SetImpl[T]) IsValid(value T) (isValid bool) {
	_, isValid = st.elementMap[value]
	return
}

func (st *SetImpl[T]) Element(value T) (element Element[T]) {
	if e, ok := st.elementMap[value]; ok {
		element = e
	}
	return
}

func (st *SetImpl[T]) Description(value T) (full string) {
	var v any = value
	if e, ok := v.(ElementFull[T]); ok {
		full = e.Description()
	}
	return
}

func PrintAsString(value any) (s string) {
	tt := t{f: value}
	s = pfmt.Sprintf("%v", tt)
	s = strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
	return
}

type t struct {
	f interface{}
}

// StringT provides a String function to a named type implementing a set
func (st *SetImpl[T]) StringT(value T) (s string) {
	// StringT is intended to be the String method of a named type implementing set.
	// if StringT method code would somehow invoke the T.String method again,
	// this will cause infinite recursion and stack overflow panic.
	var e Element[T]
	var ok bool
	if e, ok = st.elementMap[value]; ok {
		// e is Element interface, likely set.Element runtime type.
		// set.Element.String() returns a value from a type string, ie. type T
		// is not involved.
		// This ensures no T-type String-function recursion.
		s = e.String()
	} else {
		// converting T to string may recurse if printf %v or %s verbs are used.
		// ie. printf will re-enter this method and exhaust the stack.
		// all that is known about T is that it is a comparable type
		//p = string(value)

		s = "?\x27" + PrintAsString(value) + "\x27"
	}
	return
}

func (st *SetImpl[T]) String() (s string) {
	var t T
	s = pfmt.Sprintf("%T:%d", t, len(st.elementMap))
	return
}
