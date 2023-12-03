/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import "github.com/haraldrudell/parl/iters"

// SetID0 is a set of elements set.Element[T] where element type E is accessible
type SetID0[T comparable, E any] struct {
	Set0[T, E]
	fTE        func(comparable T) (ep *E)
	epIterator func() (iterator iters.Iterator[*E])
}

// NewSetID returns a set of elements set.Element[T] where element type E is accessible
func NewSetID[T comparable, E any](elements []E) (set SetID[T, E]) {
	s := SetID0[T, E]{}
	NewSetFieldp(elements, &s.Set0, &s.fTE, &s.epIterator)
	return &s
}

// Element returns the element representation for value or
// nil if value is not an element of the set
func (s *SetID0[T, E]) Element(value T) (element *E) { return s.fTE(value) }

// Iterator allows iteration over all set elements
func (s *SetID0[T, E]) EIterator() (iterator iters.Iterator[*E]) { return s.epIterator() }
