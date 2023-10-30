/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

// Iterator allows traversal of values.
//
// Iterators in the iters package are thread-safe but generally,
// thread-safety depends on the iterator implementation used.
//
// the ForIterator interface is optimized for use in the Go “for” clause
//
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
//
// ForIterator is an iterator optimized for the Init and Condition
// statements of a Go “for” clause
//
// Usage:
//
//	for i, iterator := NewIterator().Init(); iterator.Cond(&i); {
//	  println(i)
//
//	var err error
//	for i, iterator := NewIterator().Init(); iterator.Cond(&i, &err); {
//	}
//	if err != nil {
//
//	var err error
//	var iterator = NewIterator()
//	for i, iterator := NewIterator().Init(); iterator.Cond(&i); {
//	}
//	if err := iterator.Cancel() {
//	  println(i)
type Iterator[T any] interface {
	// Init implements the right-hand side of a short variable declaration in
	// the init statement for a Go “for” clause
	Init() (iterationVariable T, iterator Iterator[T])
	// Cond implements the condition statement of a Go “for” clause
	//   - the iterationVariable is updated by being provided as a pointer.
	//     iterationVariable cannot be nil
	//   - errp is an optional error pointer receiving any errors during iterator execution
	//   - condition is true if iterationVariable was assigned a value and the iteration should continue
	Cond(iterationVariablep *T, errp ...*error) (condition bool)
	// Next advances to next item and returns it
	//	- if the next item does exist, value is valid and hasValue is true
	//	- if no next item exists, value is the data type zero-value and hasValue is false
	Next() (value T, hasValue bool)
	// Same returns the same value again.
	//	- If a value does exist, it is returned in value and hasValue is true
	//	- If a value does not exist, the data type zero-value is returned and hasValue is false
	//	- If Next or Cond has not been invoked, Same first advances to the first item
	Same() (value T, hasValue bool)
	// Cancel release resources for this iterator
	//	- returns the first error that occurred during iteration if any
	//	- Not every iterator requires a Cancel invocation
	Cancel(errp ...*error) (err error)
	// // HasNext advances to next item and returns hasValue true if this next item does exists
	// HasNext() (hasValue bool)
	// // NextValue advances to next item and returns it.
	// //	- If no next value exists, the data type zero-value is returned
	// NextValue() (value T)
	// // Has returns true if Same() or SameValue will return items.
	// //	- If Next, FindNext or HasNext have not been invoked,
	// //		Has first advances to the first item
	// Has() (hasValue bool)
	// // SameValue returns the same value again
	// //	- If a value does not exist, the data type zero-value is returned
	// //	- If Next, FindNext or HasNext have not been invoked,
	// //		SameValue first advances to the first item
	// SameValue() (value T)
}
