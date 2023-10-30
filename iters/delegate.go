/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

const (
	// IsSame indicates to Delegate.Next that
	// this is a Same-type incovation
	IsSame NextAction = 0
	// IsNext indicates to Delegate.Next that
	// this is a Next-type incovation
	IsNext NextAction = 1
)

// NextAction is a unique named type that indicates whether
// the next or the same value again is sought by Delegate.Next
//   - IsSame IsNext
type NextAction uint8

// DelegateAction finds the next or the same value.
//   - isSame == IsSame returns the same value again,
//     finding the first value if a value has yet to be retrieved
//   - isSame == IsNext find the next val;ue if one exists
//   - value is the sought value or the T type’s zero-value if no value exists.
//   - hasValue indicates whether value was assigned a T value.
type DelegateAction[T any] func(isSame NextAction) (value T, hasValue bool)

// Delegate defines the methods that an iterator implementation must implement
// to use iterator.Delegator
//   - Iterator methods:
//     Next HasNext NextValue
//     Same Has SameValue
//     Cancel
//   - Delegator Methods:
//     Next HasNext NextValue
//     Same Has SameValue
//   - Delegate Methods
//     Next Cancel
type Delegate[T any] interface {
	// Next finds the next or the same value.
	// isSame indicates what value is sought.
	// value is the sought value or the T type’s zero-value if no value exists.
	// hasValue indicates whether value was assigned a T value.
	Action(isSame NextAction) (value T, hasValue bool)
	// Cancel indicates to the iterator implementation that iteration has concluded
	// and resources should be released.
	Cancel() (err error)
}
