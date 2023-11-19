/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

// Iterator allows traversal of values.
//   - iterate over an unimported type using an iterator returning
//     a derived value such as an interface type
//   - convert a slice value-by-value, ie. return interface-type values from
//     a struct slice
//   - iterate over pointers to slice values
//   - iterate over function
//   - obtain errors occurring in the iterator or release iterator resources
//
// The iterator interface is optimized for use in the Go “for” clause.
// Iterators in the iters package are thread-safe but
// thread-safety depends on iterator implementation.
//
// Usage in Go “for” clause:
//
//	for i, iterator := NewIterator().Init(); iterator.Cond(&i); {
//	  println(i)
//
//	var err error
//	for i, iterator := NewIterator().Init(); iterator.Cond(&i, &err); {
//	  println(i)
//	}
//	if err != nil {
//
//	var err error
//	var iterator = NewIterator()
//	for i := 0; iterator.Cond(&i); {
//	  println(i)
//	}
//	if err = iterator.Cancel(); err != nil {
type Iterator[T any] interface {
	// Init implements the right-hand side of a short variable declaration in
	// the init statement of a Go “for” clause
	//
	//		for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
	//	   // i is pointer to slice element
	Init() (iterationVariable T, iterator Iterator[T])
	// Cond implements the condition statement of a Go “for” clause
	//   - condition is true if iterationVariable was assigned a value and the iteration should continue
	//   - the iterationVariable is updated by being provided as a pointer.
	//     iterationVariable cannot be nil
	//   - errp is an optional error pointer receiving any errors during iterator execution
	//
	// Usage:
	//
	//  for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
	//    // i is pointer to slice element
	Cond(iterationVariablep *T, errp ...*error) (condition bool)
	// Next advances to next item and returns it
	//	- if hasValue true, value contains the next value
	//	- otherwise, no more items exist and value is the data type zero-value
	Next() (value T, hasValue bool)
	// Same returns the same value again
	//	- if hasValue true, value is valid
	//	- otherwise, no more items exist and value is the data type zero-value
	//	- If Next or Cond has not been invoked, Same first advances to the first item
	Same() (value T, hasValue bool)
	// Cancel stops an iteration
	//	- after Cancel invocation, Cond, Next and Same indicate no value available
	//	- Cancel returns the first error that occurred during iteration, if any
	//	- an iterator implementation may require Cancel invocation
	//		to release resources
	//	- Cancel is deferrable
	Cancel(errp ...*error) (err error)
}
