/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"errors"
	"sync"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
)

// ConverterIterator traverses another iterator and returns converted values. Thread-safe.
type ConverterIterator[K constraints.Ordered, T any] struct {
	lock sync.Mutex
	// nextCounter counts and identifies Next operations
	nextCounter int
	keyIterator Iterator[K]
	// fn receives a key and returns the corresponding value.
	// if fn returns error, it will not be invoked again.
	// if isCancel is true, this is the final invocation of fn and it should release any resource.
	// if isCancel is true, no value is sought.
	// if fn returns parl.ErrEndCallbacks, it will not be invoked again.
	fn func(key K, isCancel bool) (value T, err error)
	// fnEnded indicates that fn reurned parl.ErrEndCallbacks or other error.
	// when fnEnded is true, fn does not need isCancel invocation.
	fnEnded  bool
	value    T
	hasValue bool
	err      error
	// isCancel indicates that Cancel was invoked.
	// isCancel invalidates both value and hasValue.
	isCancel bool
}

// NewConverterIterator returns a converting iterator.
// ConverterIterator is trhread-safe and re-entrant.
func NewConverterIterator[K constraints.Ordered, T any](
	keyIterator Iterator[K], fn func(key K, isCancel bool) (value T, err error)) (iterator Iterator[T]) {
	if fn == nil {
		panic(perrors.NewPF("fn cannot be nil"))
	} else if keyIterator == nil {
		panic(perrors.NewPF("keyIterator cannot be nil"))
	}
	return &Delegator[T]{Delegate: &ConverterIterator[K, T]{
		keyIterator: keyIterator,
		fn:          fn,
	}}
}

// InitConverterIterator initializes a ConverterIterator struct.
// ConverterIterator is trhread-safe and re-entrant.
func InitConverterIterator[K constraints.Ordered, T any](
	iterp *ConverterIterator[K, T],
	keyIterator Iterator[K],
	fn func(key K, isCancel bool) (value T, err error),
) {
	if iterp == nil {
		panic(perrors.NewPF("iterator cannot be nil"))
	}
	if keyIterator == nil {
		panic(perrors.NewPF("keyIterator cannot be nil"))
	}
	if fn == nil {
		panic(perrors.NewPF("fn cannot be nil"))
	}
	iterp.lock = sync.Mutex{}
	iterp.nextCounter = 0
	iterp.keyIterator = keyIterator
	iterp.fn = fn
	iterp.fnEnded = false
	var value T
	iterp.value = value
	iterp.hasValue = false
	iterp.err = nil
	iterp.isCancel = false
}

func (iter *ConverterIterator[K, T]) Next(isSame NextAction) (value T, hasValue bool) {
	iter.lock.Lock()
	defer iter.lock.Unlock()

	// should next operation be invoked?
	if iter.isCancel || iter.fnEnded {
		return // error from fn or cancel invoked: zero-value and hasValue false
	}

	// same value when a value has previously been read
	if isSame == IsSame && iter.nextCounter != 0 {
		value = iter.value
		hasValue = iter.hasValue
		return
	}

	// next operation begins…
	nextID := iter.nextCounter
	iter.nextCounter++

	// get next key from keyIterator
	key, hasKey := iter.keyIterator.Next()
	if !hasKey {
		return // no more items return: zero-value, false
	}

	// get value for this key
	var err error
	if value, err = iter.convert(key); err != nil {
		iter.fnEnded = true
		if !errors.Is(err, cyclebreaker.ErrEndCallbacks) {
			iter.err = perrors.AppendError(iter.err, err)
		}
		var zeroValue T
		value = zeroValue
		iter.value = zeroValue
		iter.hasValue = false
		return // failed conversion error return: zero-value, false
	}

	hasValue = true
	if nextID+1 == iter.nextCounter {
		iter.value = value
		iter.hasValue = true
	}

	return // good return: value and hasValue true
}

func (iter *ConverterIterator[K, T]) convert(key K) (value T, err error) {
	iter.lock.Unlock()
	defer iter.lock.Lock()

	cyclebreaker.RecoverInvocationPanic(func() {
		value, err = iter.fn(key, false)
	}, &err)
	return
}

func (iter *ConverterIterator[K, T]) Cancel() (err error) {
	iter.lock.Lock()
	defer iter.lock.Unlock()

	// if already error, or Cancel has alredy been invoked, return same result
	if iter.isCancel {
		err = iter.err
		return
	}
	iter.isCancel = true

	// ensure fn canceled and have any errors in err
	if !iter.fnEnded {
		cyclebreaker.RecoverInvocationPanic(func() {
			var k K
			_, err = iter.fn(k, true)
		}, &err)
	}
	err = perrors.AppendError(iter.err, err)

	// cancel keyIterator, append to err
	if err2 := iter.keyIterator.Cancel(); err2 != nil {
		err = perrors.AppendError(err, err2)
	}

	// store error result
	if err != nil {
		iter.err = err
	}

	return
}
