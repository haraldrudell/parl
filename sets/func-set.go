/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"fmt"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pfmt"
	"github.com/haraldrudell/parl/plog"
)

// FuncSet is a set where a function eIDFunc extracts element ID from elements
type FuncSet[T comparable, E any] struct {
	elementMap map[T]*E
	elements   []E
	ts         []T
	eIDFunc    func(elementType *E) (elementID T)
}

// NewFunctionSet returns a set where a function eIDFunc extracts element ID from elements
func NewFunctionSet[T comparable, E any](elements []E, eIDFunc func(elementType *E) (elementID T)) (set SetID[T, E]) {
	if eIDFunc == nil {
		cyclebreaker.NilError("eIDFunc")
	}
	var m = make(map[T]*E, len(elements))
	var ts = make([]T, len(elements))
	for i := 0; i < len(elements); i++ {
		var ep = &elements[i]
		var t = eIDFunc(ep)
		if _, ok := m[t]; ok {
			var e E
			panic(perrors.ErrorfPF(
				"duplicate set-element:"+
					" type T: %T duplicate value: ‘%[1]v’"+
					" type E: %T duplicate value: ‘%s’"+
					" number of added values: %d",
				t,
				e, pfmt.NoRecurseVPrint(e),
				i,
			))
		}
		m[t] = ep
		ts[i] = t
	}
	return &FuncSet[T, E]{
		elementMap: m,
		elements:   elements,
		ts:         ts,
		eIDFunc:    eIDFunc,
	}
}

// IsValid returns whether value is part of the set
func (s *FuncSet[T, E]) IsValid(value T) (isValid bool) {
	_, isValid = s.elementMap[value]
	return
}

// Iterator allows iteration over all set elements
func (s *FuncSet[T, E]) Iterator() (iterator iters.Iterator[T]) { return iters.NewSliceIterator(s.ts) }

// StringT returns a string representation for an element of this set.
// If value is not a valid element, a fmt.Printf value is output like ?'%v'
func (s *FuncSet[T, E]) StringT(value T) (s2 string) {
	var ep, isValid = s.elementMap[value]
	if stringer, ok := any(ep).(fmt.Stringer); ok {
		s2 = stringer.String()
	} else {
		s2 = fmt.Sprintf("%v", *ep)
	}
	if isValid {
		return
	}
	s2 = fmt.Sprintf("?badID‘%v’", value)

	return
}

// Description returns a more elaborate string representation
// for an element. Description and StringT may return the same value
func (s *FuncSet[T, E]) Description(value T) (full string) {
	var ep, ok = s.elementMap[value]
	if !ok {
		return
	}

	// type assert to Element with Description method
	var elementDescription ElementDescription
	if elementDescription, ok = any(ep).(ElementDescription); !ok {
		return // not a full element return
	}

	full = elementDescription.Description()

	return
}

// Element returns the element representation for value or
// nil if value is not an element of the set
func (s *FuncSet[T, E]) Element(value T) (elementType *E) { return s.elementMap[value] }

// Iterator allows iteration over all set elements
func (s *FuncSet[T, E]) EIterator() (iterator iters.Iterator[*E]) {
	return iters.NewSlicePointerIterator(s.elements)
}

func (s *FuncSet[T, E]) String() (s2 string) {
	var e E
	s2 = plog.Sprintf("funcSet_%T:%d", e, len(s.elementMap))
	return
}
