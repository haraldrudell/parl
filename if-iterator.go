/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Iterator allows traversal of values.
// The iterators in parly.iterator are thread-safe and re-entrant, but generally, this depends
// on the iterator implementation used.
//
//	// triple-expression works for Iterator that do not require Cancel
//	for iterator := NewIterator(); iterator.HasNext(); {
//		v := iterator.SameValue()
//	}
//
//	// conditional expression can be used with all iterators
//	iterator := NewIterator()
//	for iterator.HasNext() {
//		v := iterator.SameValue()
//	}
//	if err = iterator.Cancel(); …
type Iterator[T any] interface {
	// Next advances to next item and returns it.
	// if the next item does exist, value is valid and hasValue is true.
	// if no next item exists, value is the data type zero-value and hasValue is false.
	Next() (value T, hasValue bool)
	// HasNext advances to next item and returns hasValue true if this next item does exists.
	HasNext() (hasValue bool)
	// NextValue advances to next item and returns it.
	// If no next value exists, the data type zero-value is returned.
	NextValue() (value T)
	// Same returns the same value again.
	// If a value does exist, it is returned in value and hasValue is true.
	// If a value does not exist, the data type zero-value is returned and hasValue is false.
	// If Next, FindNext or HasNext have not been invoked, Same first advances to the first item.
	Same() (value T, hasValue bool)
	// Has returns true if Same() or SameValue will return items.
	// If Next, FindNext or HasNext have not been invoked, Has first advances to the first item.
	Has() (hasValue bool)
	// SameValue returns the same value again.
	// If a value does not exist, the data type zero-value is returned.
	// If Next, FindNext or HasNext have not been invoked, SameValue first advances to the first item.
	SameValue() (value T)
	// Cancel release resources for this iterator.
	// Not every iterator requires a Cancel invocation.
	Cancel() (err error)
}
