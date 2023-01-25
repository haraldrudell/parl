/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Set provides a collection of unique elements of a particular type T that are printable,
// type convertible and have verifiable validity.
// Element represents an element of a set that has a unique value and is printable.
// SetElements provides an iterator for all elements of a set intended for SetFactory.NewSet.
package sets

import (
	"fmt"

	"github.com/haraldrudell/parl/iters"
)

// Set provides a collection of unique elements of a particular type T that are printable,
// type convertible and have verifiable validity.
// A set stores unique values without any particular order.
// The reasons for using a set over const are:
//   - set memberhip enforcement
//   - available string representation for elements
//   - additional fields or methods assigned to elements
//   - optional type safety
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
type Set[T comparable] interface {
	// IsValid returns whether value is part of this set
	IsValid(value T) (isValid bool)
	// Element returns the element representation for value or
	// nil if value is not an element of the set.
	Element(value T) (element Element[T])
	// StringT returns a string representation for an element of this set.
	// if value is not a valid element, a fmt.Printf value is output like ?'%v'
	StringT(value T) (s string)
	Description(value T) (s string)
	fmt.Stringer
}

// Element represents an element of a set that has a unique value and is printable.
// set element values are unique but not necessarily ordered.
// set.SetElement is an implementation.
type Element[T comparable] interface {
	Value() (value T)
	fmt.Stringer
}

// ElementIterator provides an iterator for all elements of a set intended for SetFactory.NewSet.
type ElementIterator[T comparable] interface {
	Elements() iters.Iterator[*Element[T]]
}

type ElementFull[T comparable] interface {
	Element[T]
	Description() (full string)
}

type SetFactory[T comparable] interface {
	// NewSet returns a set of a finite number of elements.
	// Usage:
	//
	//	const IsSame NextAction = 0
	//	type NextAction uint8
	//	func (na NextAction) String() (s string) {return nextActionSet.StringT(na)}
	//	var nextActionSet = set.NewSet(yslices.ConvertSliceToInterface[
	//	  set.Element[NextAction],
	//	  parli.Element[NextAction],
	//	]([]set.Element[NextAction]{{ValueV: IsSame, Name: "IsSame"}}))
	NewSet(elements []Element[T]) (interfaceSet Set[T])
}
