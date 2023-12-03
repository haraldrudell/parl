/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"fmt"

	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/pfmt"
	"github.com/haraldrudell/parl/plog"
)

// SliceSet is a set of elements that may not be comparable
//   - slice index int is used as untyped identity
//   - elements may be a struct with slice, map or function fields, or
//     an interface type whose dynamic type is one of those
//   - if elements have String method, this is description.
//     Otherwise the “IDn” is description
//   - —
//   - SliceSet does:
//   - — ID range checking
//   - — provides pointer-iterator
//   - a set should have:
//   - — its own unique named type
//   - — element ID literal, usually a typed const value
//   - — an iterator
//   - — a way to determine if a value is valid
type SliceSet[E any] struct {
	elements   []E
	isStringer bool
	typeName   string
}

// NewSliceSet is a slice-based set using slice index as comparable
//   - element type is therefore any
//   - if the slice changes, IDs change, too
func NewSliceSet[E any](elements []E) (set SetID[int, E]) {
	var e E
	var _, isStringer = any(e).(fmt.Stringer)
	return &SliceSet[E]{
		elements:   elements,
		typeName:   fmt.Sprintf("%T", e),
		isStringer: isStringer,
	}
}

// IsValid returns whether value is part of the set
func (s *SliceSet[E]) IsValid(elementID int) (isValid bool) {
	return elementID >= 0 && elementID < len(s.elements)
}

// Iterator allows iteration over all set elements
func (s *SliceSet[E]) Iterator() (iterator iters.Iterator[int]) {
	if len(s.elements) == 0 {
		iterator = iters.NewEmptyIterator[int]()
		return
	}
	iterator = iters.NewIntegerIterator(0, len(s.elements)-1)

	return
}

// StringT returns a string representation for an element of this set.
// If value is not a valid element, a fmt.Printf value is output like ?'%v'
func (s *SliceSet[E]) StringT(anID int) (s2 string) {
	if anID < 0 || anID >= len(s.elements) {
		s2 = fmt.Sprintf("?badID‘%d’", anID)
		return
	}
	var value = s.elements[anID]
	// StringT is intended to be the String method of a named type implementing set.
	// if StringT method code would somehow invoke the T.String method again,
	// this will cause infinite recursion and stack overflow panic.
	s2 = pfmt.NoRecurseVPrint(value)

	return
}

// Description returns a more elaborate string representation
// for an element. Description and StringT may return the same value
func (s *SliceSet[E]) Description(anID int) (full string) {
	if anID < 0 || anID >= len(s.elements) {
		return
	}
	var ep = &s.elements[anID]

	// type assert to Element with Description method
	var elementDescription ElementDescription
	var ok bool
	if elementDescription, ok = any(ep).(ElementDescription); !ok {
		return // not a full element return
	}

	full = elementDescription.Description()

	return
}

// Element returns the element representation for value or
// nil if value is not an element of the set
func (s *SliceSet[E]) Element(anID int) (elementType *E) {
	if anID < 0 || anID >= len(s.elements) {
		return
	}
	elementType = &s.elements[anID]

	return
}

// Iterator allows iteration over all set elements
func (s *SliceSet[E]) EIterator() (iterator iters.Iterator[*E]) {
	return iters.NewSlicePointerIterator(s.elements)
}

func (s *SliceSet[E]) String() (s2 string) {
	return plog.Sprintf("sliceSet_%T:%d", s.typeName, len(s.elements))
}
