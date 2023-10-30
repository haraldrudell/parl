/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

func TestFunctionIterator(t *testing.T) {
	var values = []string{"one", "two"}

	var value string
	var hasValue bool
	var zeroValue string
	var err error

	// test type that has tt.fn function that can be used with function iterator
	var tt *fnIteratorTester = newFnIteratorTester(values, t)
	// the iterator under test
	var iterator Iterator[string] = NewFunctionIterator(tt.fn)

	// NewFunctionIterator should return a value
	if iterator == nil {
		t.Error("iterator nil")
		t.FailNow()
	}

	// request IsSame value twice should:
	//	- retrieve the first value and return it
	//	- then return the same value again
	for i := 0; i <= 1; i++ {
		value, hasValue = iterator.Same()

		//hasValue should be true
		if !hasValue {
			t.Errorf("Same%d: hasValue false", i)
		}
		// value should be first value
		if value != values[0] {
			t.Errorf("Same%d value %q exp %q", i, value, values[0])
		}
	}

	// Next should return the second value
	value, hasValue = iterator.Next()
	if !hasValue {
		t.Errorf("Next hasValue false")
	}
	if value != values[1] {
		t.Errorf("Next value %q exp %q", value, values[1])
	}

	// Next should return no value
	value, hasValue = iterator.Next()
	if hasValue {
		t.Errorf("Next2 hasValue true")
	}
	if value != zeroValue {
		t.Errorf("Next2 value %q exp %q", value, zeroValue)
	}

	// cancel should not return error
	if err = iterator.Cancel(); err != nil {
		t.Errorf("Cancel err '%v'", err)
	}

	// Next after cancel should not return a value
	value, hasValue = iterator.Next()
	if hasValue {
		t.Errorf("Next3 hasValue true")
	}
	if value != zeroValue {
		t.Errorf("Next3 value %q exp %q", value, zeroValue)
	}
}

// tests interface Iterator[string]
func TestNewFunctionIterator(t *testing.T) {
	var values = []string{}
	var zeroValue string
	var messageIterpNil = "cannot be nil"

	var iterator Iterator[string]
	var value string
	var hasValue bool
	var err error

	// test type that has tt.fn function that can be used with function iterator
	var tt = newFnIteratorTester(values, t)
	iterator = NewFunctionIterator(tt.fn)

	// Same should retrieve the first value, but there isn’t one
	value, hasValue = iterator.Same()
	// hasValue should be false
	if hasValue {
		t.Error("Same hasValue true")
	}
	// value shouldbe zero-value
	if value != zeroValue {
		t.Error("Same hasValue not zeroValue")
	}

	// new with nil argument should panic
	_, err = tt.invokeNewFunctionIterator(nil)
	if err == nil || !strings.Contains(err.Error(), messageIterpNil) {
		t.Errorf("InitSliceIterator incorrect panic: '%v' exp %q", err, messageIterpNil)
	}

	// new and cancel should not return error
	iterator, err = tt.invokeNewFunctionIterator(tt.fn)
	_ = err
	err = iterator.Cancel()
	if err != nil {
		t.Errorf("Cancel2 err: '%v'", err)
	}
}

// fnIteratorTester is a fixture for testing Iterator[T] implementations
type fnIteratorTester struct {
	index  int        // index is current index in slice
	values []string   // values are ethe values provided during iteration
	t      *testing.T // testing instance
}

func newFnIteratorTester(values []string, t *testing.T) (tt *fnIteratorTester) {
	return &fnIteratorTester{
		values: values,
		t:      t,
	}
}

var _ IteratorFunction[string] = (&fnIteratorTester{}).fn

// tt.fn is a function that can be used with function iterator
func (tt *fnIteratorTester) fn(isCancel bool) (value string, err error) {
	t := tt.t

	// on cancel request
	if isCancel {
		t.Log("fn cancel")
		return
	}

	// index should be a sequence 0…

	// index beyond number of values
	if tt.index >= len(tt.values) {
		t.Log("fn ErrEndCallbacks")
		err = cyclebreaker.ErrEndCallbacks
		_ = errors.As
		return
	}

	// return value
	value = tt.values[tt.index]
	tt.index++
	t.Logf("fn value: %q", value)

	return
}
func (tt *fnIteratorTester) invokeNewFunctionIterator(fn IteratorFunction[string]) (
	iterator Iterator[string],
	err error,
) {
	defer func() {
		v := recover()
		if v == nil {
			return
		}
		err = v.(error)
	}()

	iterator = NewFunctionIterator[string](fn)
	return
}
