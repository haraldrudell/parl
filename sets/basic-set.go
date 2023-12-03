/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfmt"
	"github.com/haraldrudell/parl/plog"
)

// BasicSet elements are basic-type comparable
type BasicSet[E comparable] struct {
	elementMap map[E]struct{}
	elements   []E
}

var emptyStuct struct{}

// NewBasicSet returns a set of basic-type comparable elements
func NewBasicSet[E comparable](elements []E) (set Set[E]) {
	var m = make(map[E]struct{}, len(elements))
	for i, e := range elements {
		if _, ok := m[e]; ok {
			panic(perrors.ErrorfPF(
				"duplicate set-element:"+
					" type E: %T duplicate value: ‘%s’"+
					" number of added values: %d",
				e, pfmt.NoRecurseVPrint(e),
				i,
			))
		}
		m[e] = emptyStuct
	}
	return &BasicSet[E]{
		elementMap: m,
		elements:   elements,
	}
}

// IsValid returns whether value is part of the set
func (s *BasicSet[E]) IsValid(value E) (isValid bool) {
	_, isValid = s.elementMap[value]
	return
}

// Iterator allows iteration over all set elements
func (s *BasicSet[E]) Iterator() (iterator iters.Iterator[E]) {
	return iters.NewSliceIterator(s.elements)
}

// StringT returns a string representation for an element of this set.
// If value is not a valid element, a fmt.Printf value is output like ?'%v'
func (s *BasicSet[E]) StringT(value E) (s2 string) {
	// StringT is intended to be the String method of a named type implementing set.
	// if StringT method code would somehow invoke the T.String method again,
	// this will cause infinite recursion and stack overflow panic.
	s2 = pfmt.NoRecurseVPrint(value)
	if _, ok := s.elementMap[value]; ok {
		return
	}
	s2 = "?‘" + s2 + "’"

	return
}

// Description returns a more elaborate string representation
// for an element. Description and StringT may return the same value
func (s *BasicSet[E]) Description(value E) (full string) {

	// type assert to Element with Description method
	var elementDescription ElementDescription
	var ok bool
	if elementDescription, ok = any(&value).(ElementDescription); !ok {
		return // not a full element return
	}

	full = elementDescription.Description()

	return
}

func (s *BasicSet[E]) String() (s2 string) {
	var e E
	s2 = plog.Sprintf("basicSet_%T:%d", e, len(s.elementMap))
	return
}
