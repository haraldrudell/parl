/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package iters

import (
	"testing"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"github.com/haraldrudell/parl/perrors"
)

func TestFunction(t *testing.T) {
	var value1, value2 = 1, 2
	var values = []int{value1, value2}

	var err error
	var value, zeroValue, iterationVariable int
	var hasValue, condition bool
	var actIterator Iterator[int]

	var iterator Iterator[int]
	var reset = func() {
		iterator = NewFunctionIterator(newHasFunc(values).IteratorFunction)
	}
	// Init() Cond() Next() Same() Cancel()
	var _ Slice[int]

	// Init should return zero value and iterator
	reset()
	iterationVariable, actIterator = iterator.Init()
	if iterationVariable != zeroValue {
		t.Errorf("Init iterationVariable %d exp %d", iterationVariable, zeroValue)
	}
	if actIterator != iterator {
		t.Error("Init iterator bad")
	}

	// Cond should return true and update value
	reset()
	value = zeroValue
	condition = iterator.Cond(&value)
	if !condition {
		t.Error("Cond condition false")
	}
	if value != value1 {
		t.Errorf("Cond value %d exp %d", value, value1)
	}

	// Next should return first value
	reset()
	value, hasValue = iterator.Next()
	if !hasValue {
		t.Error("Next hasValue false")
	}
	if value != value1 {
		t.Errorf("Next value %d exp %d", value, value1)
	}

	// Cancel should return no error
	reset()
	err = iterator.Cancel()
	if err != nil {
		t.Errorf("Cancel err: %s", perrors.Short(err))
	}

	// CondCond should return second value
	reset()
	condition = iterator.Cond(&value)
	_ = condition
	value = zeroValue
	condition = iterator.Cond(&value)
	if !condition {
		t.Error("CondCond condition false")
	}
	if value != value2 {
		t.Errorf("Cond value %d exp %d", value, value2)
	}

	// CondCondCond should return no value
	// request IsSame value twice should:
	//   - retrieve the first value and return it
	//   - then return the same value again
	//
	// Cond should return true and update value
	reset()
	condition = iterator.Cond(&value)
	_ = condition
	condition = iterator.Cond(&value)
	_ = condition
	value = zeroValue
	condition = iterator.Cond(&value)
	if condition {
		t.Error("CondCondCond condition true")
	}
	if value != zeroValue {
		t.Errorf("CondCondCond value %d exp %d", value, zeroValue)
	}

	// NextNext should return second value
	reset()
	value, hasValue = iterator.Next()
	_ = value
	_ = hasValue
	value, hasValue = iterator.Next()
	if !hasValue {
		t.Error("NextNext hasValue false")
	}
	if value != value2 {
		t.Errorf("NextNext value %d exp %d", value, value2)
	}

	// NextNextNext should return no value
	reset()
	value, hasValue = iterator.Next()
	_ = value
	_ = hasValue
	value, hasValue = iterator.Next()
	_ = value
	_ = hasValue
	value, hasValue = iterator.Next()
	if hasValue {
		t.Error("NextNextNext hasValue true")
	}
	if value != zeroValue {
		t.Errorf("NextNextNext value %d exp %d", value, zeroValue)
	}

	// CancelCond should return false
	reset()
	err = iterator.Cancel()
	_ = err
	condition = iterator.Cond(&value)
	if condition {
		t.Error("CancelCond condition true")
	}

	// CancelNext should return no value
	reset()
	err = iterator.Cancel()
	_ = err
	value, hasValue = iterator.Next()
	if hasValue {
		t.Error("CancelNext hasValue true")
	}
	if value != zeroValue {
		t.Errorf("CancelNext value %d exp %d", value, zeroValue)
	}
}

type hasFunc[T any] struct {
	slice []T
	index int
}

func newHasFunc[T any](slice []T) (s *hasFunc[T]) { return &hasFunc[T]{slice: slice} }
func (s *hasFunc[T]) IteratorFunction(isCancel bool) (value T, err error) {
	if isCancel {
		return
	} else if s.index >= len(s.slice) {
		err = cyclebreaker.ErrEndCallbacks
		return
	}
	value = s.slice[s.index]
	s.index++

	return
}
