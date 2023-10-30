/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import "github.com/haraldrudell/parl/perrors"

// Delegator implements the value methods required by the [Iterator] interface
//   - Next HasNext NextValue
//     Same Has SameValue
//   - the delegate provides DelegateAction[T] function
//   - — delegate must be thread-safe
//   - methods provided by Delegator: Has HasNext NextValue SameValue Same
//   - delegate and Delegator combined fully implement the Iterator interface
//   - Delegator is thread-safe
type Delegator[T any] struct {
	delegateAction DelegateAction[T]
}

// NewDelegator returns an [Iterator] based on a Delegate iterator implementation.
// Delegator is thread-safe if its delegate is thread-safe.
func NewDelegator[T any](delegateAction DelegateAction[T]) (delegator *Delegator[T]) {
	if delegateAction == nil {
		perrors.NewPF("delegateAction cannot be nil")
	}

	return &Delegator[T]{delegateAction: delegateAction}
}

// Next advances to next item and returns it.
//   - if the next item does exist, value is valid and hasValue is true.
//   - if no next item exists, value is the data type zero-value and hasValue is false.
func (d *Delegator[T]) Next() (value T, hasValue bool) { return d.delegateAction(IsNext) }

// HasNext advances to next item and returns hasValue true if this next item does exists.
func (d *Delegator[T]) HasNext() (hasValue bool) {
	_, hasValue = d.Next()
	return
}

// NextValue advances to next item and returns it
//   - If no next value exists, the data type zero-value is returned.
func (d *Delegator[T]) NextValue() (value T) {
	value, _ = d.Next()
	return
}

// Same returns the same value again
//   - If a value does exist, it is returned in value and hasValue is true.
//   - If a value does not exist, the data type zero-value is returned and hasValue is false.
//   - If Next, FindNext or HasNext have not been invoked, Same first advances to the first item.
func (d *Delegator[T]) Same() (value T, hasValue bool) { return d.delegateAction(IsSame) }

// Has returns true if Same() or SameValue will return items
//   - If Next, FindNext or HasNext have not been invoked, Has first advances to the first item.
func (d *Delegator[T]) Has() (hasValue bool) {
	_, hasValue = d.Same()
	return
}

// SameValue returns the same value again
//   - If a value does not exist, the data type zero-value is returned.
//   - If Next, FindNext or HasNext have not been invoked, SameValue first advances to the first item.
func (d *Delegator[T]) SameValue() (value T) {
	value, _ = d.Same()
	return
}
