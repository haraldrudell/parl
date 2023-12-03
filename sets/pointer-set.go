/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfmt"
	"github.com/haraldrudell/parl/plog"
)

// PointerSet is a set where pointer-value provides comparable identity
type PointerSet[E any] struct {
	elementMap map[*E]struct{}
	elements   []*E
}

// NewPointerSet returns a set where pointer-value provides comparable identity
//   - on every process launch, element identity changes
func NewPointerSet[E any](elements []*E) (set Set[*E]) {
	var m = make(map[*E]struct{}, len(elements))
	for i, ep := range elements {
		if ep == nil {
			panic(perrors.ErrorfPF(
				"element-pointer cannot be nil:"+
					" pointer-type E: %T index: %d",
				ep, i,
			))
		}
		if _, ok := m[ep]; ok {
			panic(perrors.ErrorfPF(
				"duplicate set-element:"+
					" pointer-type E: %T duplicate value: 0x%x"+
					" number of added values: %d",
				ep, cyclebreaker.Uintptr(ep),
				i,
			))
		}
		m[ep] = emptyStuct
	}
	return &PointerSet[E]{
		elementMap: m,
		elements:   elements,
	}
}

// IsValid returns whether value is part of the set
func (s *PointerSet[E]) IsValid(value *E) (isValid bool) {
	_, isValid = s.elementMap[value]
	return
}

// Iterator allows iteration over all set elements
func (s *PointerSet[E]) Iterator() (iterator iters.Iterator[*E]) {
	return iters.NewSliceIterator(s.elements)
}

// StringT returns a string representation for an element of this set.
// If value is not a valid element, a fmt.Printf value is output like ?'%v'
func (s *PointerSet[E]) StringT(value *E) (s2 string) {
	// StringT is intended to be the String method of a named type implementing set.
	// if StringT method code would somehow invoke the T.String method again,
	// this will cause infinite recursion and stack overflow panic.
	if _, ok := s.elementMap[value]; ok {
		s2 = pfmt.NoRecurseVPrint(*value)
		return
	}
	s2 = "?‘" + s2 + "’"

	return
}

// Description returns a more elaborate string representation
// for an element. Description and StringT may return the same value
func (s *PointerSet[E]) Description(value *E) (full string) {

	// type assert to Element with Description method
	var elementDescription ElementDescription
	var ok bool
	if elementDescription, ok = any(value).(ElementDescription); !ok {
		return // not a full element return
	}

	full = elementDescription.Description()

	return
}

func (s *PointerSet[E]) String() (s2 string) {
	var e E
	s2 = plog.Sprintf("pointerSet_%T:%d", e, len(s.elementMap))
	return
}
