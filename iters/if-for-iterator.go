/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

// ForIterator is a simple iterator that cannot return errors
// or require resources to be released
type ForIterator[T any] interface {
	// Init implements the right-hand side of a short variable declaration in
	// the init statement of a Go “for” clause
	//
	//		for i, iterator := iters.NewSlicePointerIterator(someSlice).Init(); iterator.Cond(&i); {
	//	   // i is pointer to slice element
	Init() (iterationVariable T, iterator ForIterator[T])
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
}
