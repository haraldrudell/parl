/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

// EmptyIterator is an iterator with no values. thread-safe.
type EmptyIterator[T any] struct{}

// NewEmptyIterator returns an empty iterator of values type T.
//   - EmptyIterator is thread-safe.
func NewEmptyIterator[T any]() (iterator Iterator[T]) { return &EmptyIterator[T]{} }
func (i *EmptyIterator[T]) Init() (iterationVariable T, iterator Iterator[T]) {
	iterator = i
	return
}
func (i *EmptyIterator[T]) Cond(iterationVariablep *T, errp ...*error) (condition bool) { return }
func (i *EmptyIterator[T]) Next() (value T, hasValue bool)                              { return }
func (i *EmptyIterator[T]) HasNext() (ok bool)                                          { return }
func (i *EmptyIterator[T]) NextValue() (value T)                                        { return }
func (i *EmptyIterator[T]) Same() (value T, hasValue bool)                              { return }
func (i *EmptyIterator[T]) Has() (hasValue bool)                                        { return }
func (i *EmptyIterator[T]) SameValue() (value T)                                        { return }
func (i *EmptyIterator[T]) Cancel(errp ...*error) (err error)                           { return }
