/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sets

import (
	"fmt"

	"github.com/haraldrudell/parl/iters"
)

// Set is a collection of elements.
//
// A set has:
//   - a unique named type for elements that is comparable
//   - element values and literals providing access to set functions:
//   - — iterator over all elements
//   - — IsValid method determining if the element is valid or zero-value
//   - — String method making the element printable
//   - — for elements with element ID, access to the element type
//
// A set element has:
//   - element or element ID literals, typically typed constants
//   - series of element or element ID literals are often
//     typed iota-based constants
//   - comparable mechanic:
//   - — an incomparable element must have a comparable element ID
//   - —	element ID may also otherwise be used to refer to an element
//   - — elements or element IDs may also be ordered
//
// benefits of using sets include:
//   - Iterator method iterating over all elements
//   - IsValid method verifying whether value is valid
//   - String method ensuring elements to be printable
//   - make incomparable elements comparable and iterable
//   - referring to a complex element any-type using
//     a basic-type comparable element ID
//   - type safety by:
//   - — distinguishing the set from all other types and values
//   - — using elements or element IDs as typed function arguments
//
// composite or incomparable elements
//   - If the element type is incomparable,
//     the element should have a basic-type ID for which
//     literals exist:
//   - — Integer, Complex string float boolean
//   - — slice index or pointers can be used as identity and element iD,
//     but this complicates element ID literals or storing of element IDs
//   - incomparables are slice, map, function, and structs or interface
//     dynamic types of those. In particular, a byte slice is incomparable
//   - an element must at least have a derived comparable ID type
//   - a method can be used to obtain a comparable ID from a
//     composite or incomparable type
//
// complex elements
//   - similarly, if element type is large, it may benefit from being referred
//     by an element ID
//
// Usage:
//
//	const IsSame NextAction = 0
//	type NextAction uint8
//	func (na NextAction) String() (s string) {return nextActionSet.StringT(na)}
//	var nextActionSet = set.NewSet(yslices.ConvertSliceToInterface[
//	  set.Element[NextAction],
//	  parli.Element[NextAction],
//	]([]set.Element[NextAction]{{ValueV: IsSame, Name: "IsSame"}}))
type Set[E comparable] interface {
	// IsValid returns whether value is part of the set
	IsValid(value E) (isValid bool)
	// Iterator allows iteration over all set elements
	Iterator() (iterator iters.Iterator[E])
	// StringT returns a string representation for an element of this set.
	// If value is not a valid element, a fmt.Printf value is output like ?'%v'
	StringT(value E) (s string)
	// Description returns a more elaborate string representation
	// for an element. Description and StringT may return the same value
	Description(value E) (s string)
	fmt.Stringer
}

type SetID[T comparable, E any] interface {
	Set[T]
	// Element returns the element representation for value or
	// nil if value is not an element of the set
	Element(value T) (elementType *E)
	// Iterator allows iteration over all set elements
	EIterator() (iterator iters.Iterator[*E])
}

// Element represents an element of a set that has a unique value and is printable.
// set element values are unique but not necessarily ordered.
// set.SetElement is an implementation.
type Element[T comparable] interface {
	Value() (value T)
	fmt.Stringer
}

type ElementDescription interface {
	Description() (full string)
}
