/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

// Delegator adds delegating methods that implements the Iterator interface.
// Delegator is thread-safe if its delegate is thread-safe.
type Delegator[T any] struct {
	Delegate[T]
}

// NewDelegator returns a parli.Iterator based on a Delegate iterator implementation.
// Delegator is thread-safe if its delegate is thread-safe.
func NewDelegator[T any](delegate Delegate[T]) (iter Iterator[T]) {
	return &Delegator[T]{Delegate: delegate}
}

func (iter *Delegator[T]) Next() (value T, hasValue bool) {
	return iter.Delegate.Next(IsNext)
}

func (iter *Delegator[T]) HasNext() (ok bool) {
	_, ok = iter.Next()
	return
}

func (iter *Delegator[T]) NextValue() (value T) {
	value, _ = iter.Next()
	return
}

func (iter *Delegator[T]) Same() (value T, hasValue bool) {
	return iter.Delegate.Next(IsSame)
}

func (iter *Delegator[T]) Has() (hasValue bool) {
	_, hasValue = iter.Same()
	return
}

func (iter *Delegator[T]) SameValue() (value T) {
	value, _ = iter.Same()
	return
}
