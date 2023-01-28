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
	slice := []string{"one", "two"}
	//messageFnNil := "fn cannot be nil"

	var value string
	var hasValue bool
	var zeroValue string
	var iter FunctionIterator[string]
	fn := func(index int) (value string, err error) {
		if index == FunctionIteratorCancel {
			t.Log("fn cancel")
			return
		}
		if index >= len(slice) {
			t.Log("fn ErrEndCallbacks")
			err = cyclebreaker.ErrEndCallbacks
			_ = errors.As
			return
		}
		value = slice[index]
		t.Logf("fn value: %q", value)
		return
	}
	var err error

	InitFunctionIterator(&iter, fn)

	// Same twice
	for i := 0; i <= 1; i++ {
		if value, hasValue = iter.Next(IsSame); !hasValue {
			t.Errorf("Same%d hasValue false", i)
		}
		if value != slice[0] {
			t.Errorf("Same%d value %q exp %q", i, value, slice[0])

		}
	}

	// Next
	if value, hasValue = iter.Next(IsNext); !hasValue {
		t.Errorf("Next hasValue false")
	}
	if value != slice[1] {
		t.Errorf("Next value %q exp %q", value, slice[1])
	}
	if value, hasValue = iter.Next(IsNext); hasValue {
		t.Errorf("Next2 hasValue true")
	}
	if value != zeroValue {
		t.Errorf("Next2 value %q exp %q", value, zeroValue)
	}

	if err = iter.Cancel(); err != nil {
		t.Errorf("Cancel err '%v'", err)
	}

	if value, hasValue = iter.Next(IsNext); hasValue {
		t.Errorf("Next3 hasValue true")
	}
	if value != zeroValue {
		t.Errorf("Next3 value %q exp %q", value, zeroValue)
	}
}

func TestNewFunctionIterator(t *testing.T) {
	slice := []string{}
	var zeroValue string
	messageIterpNil := "cannot be nil"
	fn := func(index int) (value string, err error) {
		if index == FunctionIteratorCancel {
			return
		}
		if index >= len(slice) {
			err = cyclebreaker.ErrEndCallbacks
			return
		}
		value = slice[index]
		return
	}

	var iter Iterator[string]
	var value string
	var hasValue bool
	var err error

	iter = NewFunctionIterator(fn)

	if value, hasValue = iter.Same(); hasValue {
		t.Error("Same hasValue true")
	}
	if value != zeroValue {
		t.Error("Same hasValue not zeroValue")
	}

	err = nil
	cyclebreaker.RecoverInvocationPanic(func() {
		NewFunctionIterator[string](nil)
	}, &err)
	if err == nil || !strings.Contains(err.Error(), messageIterpNil) {
		t.Errorf("InitSliceIterator incorrect panic: '%v' exp %q", err, messageIterpNil)
	}

	if err = NewFunctionIterator(fn).Cancel(); err != nil {
		t.Errorf("Cancel2 err: '%v'", err)
	}
}

func TestInitFunctionIterator(t *testing.T) {
	slice := []string{"one", "two"}
	messageIterpNil := "cannot be nil"
	fn := func(index int) (value string, err error) {
		if index == FunctionIteratorCancel {
			return
		}
		if index >= len(slice) {
			err = cyclebreaker.ErrEndCallbacks
			return
		}
		value = slice[index]
		return
	}

	var iter FunctionIterator[string]
	var value string
	var hasValue bool
	var iterpNil *FunctionIterator[string]
	var err error

	InitFunctionIterator(&iter, fn)

	if value, hasValue = iter.Next(IsSame); !hasValue {
		t.Error("Same hasValue false")
	}
	if value != slice[0] {
		t.Errorf("Same value %q exp %q", value, slice[0])
	}

	err = nil
	cyclebreaker.RecoverInvocationPanic(func() {
		InitFunctionIterator(iterpNil, fn)
	}, &err)
	if err == nil || !strings.Contains(err.Error(), messageIterpNil) {
		t.Errorf("InitSliceIterator incorrect panic: '%v' exp %q", err, messageIterpNil)
	}

	err = nil
	cyclebreaker.RecoverInvocationPanic(func() {
		InitFunctionIterator(&iter, nil)
	}, &err)
	if err == nil || !strings.Contains(err.Error(), messageIterpNil) {
		t.Errorf("InitSliceIterator incorrect panic: '%v' exp %q", err, messageIterpNil)
	}
}
