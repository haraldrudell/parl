/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

// Empty is an iterator with no values. thread-safe.
type Empty[T any] struct{}

// NewEmptyIterator returns an empty iterator of values type T.
//   - EmptyIterator is thread-safe.
func NewEmptyIterator[T any]() (iterator Iterator[T]) { return &Empty[T]{} }
func (i *Empty[T]) Init() (iterationVariable T, iterator Iterator[T]) {
	iterator = i
	return
}
func (i *Empty[T]) Cond(iterationVariablep *T, errp ...*error) (condition bool) { return }
func (i *Empty[T]) Next() (value T, hasValue bool)                              { return }
func (i *Empty[T]) Same() (value T, hasValue bool)                              { return }
func (i *Empty[T]) Cancel(errp ...*error) (err error)                           { return }
