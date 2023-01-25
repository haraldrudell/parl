/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

// Delegate defines the methods that an iterator implementation must implement
// to use iterator.Delegator
type Delegate[T any] interface {
	// Next finds the next or the same value.
	// isSame indicates what value is sought.
	// value is the sought value or the T type’s zero-value if no value exists.
	// hasValue indicates whether value was assigned a T value.
	Next(isSame NextAction) (value T, hasValue bool)
	// Cancel indicates to the iterator implementation that iteration has concluded
	// and resources should be released.
	Cancel() (err error)
}
