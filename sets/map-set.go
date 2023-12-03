/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"fmt"
	"slices"

	"github.com/haraldrudell/parl/iters"
	"github.com/haraldrudell/parl/pfmt"
	"github.com/haraldrudell/parl/plog"
	"golang.org/x/exp/constraints"
)

type MapSet[T constraints.Ordered, E any] struct {
	elementMap map[T]E
	elements   []sorter[T, E]
}

type sorter[T constraints.Ordered, E any] struct {
	t T
	e E
}

func NewMapSet[T constraints.Ordered, E any](m map[T]E) (set SetID[T, E]) {

	// sort elements
	//	- no pointers is 1 allocation
	var sortable = make([]sorter[T, E], len(m))
	var i int
	for t, e := range m {
		sortable[i] = sorter[T, E]{t: t, e: e}
		i++
	}
	slices.SortFunc(sortable, sorterCmp[T, E])

	return &MapSet[T, E]{
		elementMap: m,
		elements:   sortable,
	}
}

// IsValid returns whether value is part of the set
func (s *MapSet[T, E]) IsValid(value T) (isValid bool) {
	_, isValid = s.elementMap[value]
	return
}

// Iterator allows iteration over all set elements
func (s *MapSet[T, E]) Iterator() (iterator iters.Iterator[T]) {
	return iters.NewSimpleConverterIterator(
		iters.NewSlicePointerIterator(s.elements),
		s.convertT,
	)
}

// StringT returns a string representation for an element of this set.
// If value is not a valid element, a fmt.Printf value is output like ?'%v'
func (s *MapSet[T, E]) StringT(value T) (s2 string) {
	var e, ok = s.elementMap[value]
	if !ok {
		s2 = fmt.Sprintf("?badID‘%v’", value)
		return
	}
	// StringT is intended to be the String method of a named type implementing set.
	// if StringT method code would somehow invoke the T.String method again,
	// this will cause infinite recursion and stack overflow panic.
	s2 = pfmt.NoRecurseVPrint(e)

	return
}

// Description returns a more elaborate string representation
// for an element. Description and StringT may return the same value
func (s *MapSet[T, E]) Description(value T) (full string) {
	var e, ok = s.elementMap[value]
	if !ok {
		return
	}
	// type assert to Element with Description method
	var elementFull ElementDescription
	if elementFull, ok = any(&e).(ElementDescription); !ok {
		return // not a full element return
	}

	full = elementFull.Description()

	return
}

// Element returns the element representation for value or
// nil if value is not an element of the set
func (s *MapSet[T, E]) Element(value T) (elementType *E) {
	var e, ok = s.elementMap[value]
	if !ok {
		return
	}
	elementType = &e

	return
}

// Iterator allows iteration over all set elements
func (s *MapSet[T, E]) EIterator() (iterator iters.Iterator[*E]) {
	return iters.NewSimpleConverterIterator(
		iters.NewSlicePointerIterator(s.elements),
		s.convertE,
	)

}

func (s *MapSet[T, E]) convertT(tePair *sorter[T, E]) (value T) { return tePair.t }

func (s *MapSet[T, E]) convertE(tePair *sorter[T, E]) (elementType *E) { return &tePair.e }

func (s *MapSet[T, E]) String() (s2 string) {
	var e E
	s2 = plog.Sprintf("mapSet_%T:%d", e, len(s.elementMap))
	return
}

func sorterCmp[T constraints.Ordered, E any](a, b sorter[T, E]) (result int) {
	if a.t < b.t {
		return -1
	} else if a.t > b.t {
		return 1
	}
	return
}
