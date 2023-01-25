/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

// EmptyIterator is an iterator with no values. thread-safe.
type EmptyIterator[T any] struct{}

// NewEmptyIterator returns an empty iterator of values type T.
// EmptyIterator is thread-safe.
func NewEmptyIterator[T any]() (iterator Iterator[T]) {
	return &EmptyIterator[T]{}
}

func (iter *EmptyIterator[T]) Next() (value T, hasValue bool) { return }
func (iter *EmptyIterator[T]) HasNext() (ok bool)             { return }
func (iter *EmptyIterator[T]) NextValue() (value T)           { return }
func (iter *EmptyIterator[T]) Same() (value T, hasValue bool) { return }
func (iter *EmptyIterator[T]) Has() (hasValue bool)           { return }
func (iter *EmptyIterator[T]) SameValue() (value T)           { return }
func (iter *EmptyIterator[T]) Cancel() (err error)            { return }
