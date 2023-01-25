/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestConverterIterator(t *testing.T) {
	valueK := "value"
	valueT := 17
	iterErr := errors.New("iter-err")
	errBadK := errors.New("bad K")
	errFnShouldFail := errors.New("fnShouldError")

	var err error
	var iter Iterator[int]
	var slice = []string{valueK}
	var actualT int
	var hasValue bool
	var zeroValueT int
	var fnShouldError bool
	fn := func(key string, isCancel bool) (value int, err error) {
		if fnShouldError {
			fnShouldError = false
			err = errFnShouldFail
			return
		}
		if isCancel {
			return
		}
		if key != valueK {
			err = errBadK
			return
		}
		value = valueT
		return
	}

	// test methods

	// Next
	iter = NewConverterIterator(NewSliceIterator(slice), fn)
	// Next1: exhaust keys
	if actualT, hasValue = iter.Next(); !hasValue {
		t.Error("Next1 hasValue false")
	}
	if actualT != valueT {
		t.Errorf("Next1 %d exp: %d", actualT, valueT)
	}
	// Next2: exhaust keys
	if actualT, hasValue = iter.Next(); hasValue {
		t.Error("Next2 hasValue true")
	}
	if actualT != zeroValueT {
		t.Errorf("Next2 %d exp: %d", actualT, zeroValueT)
	}
	// Next3: fn error
	fnShouldError = true
	iter = NewConverterIterator(NewSliceIterator(slice), fn)
	if actualT, hasValue = iter.Next(); hasValue {
		t.Error("Next3 hasValue true")
	}
	if actualT != zeroValueT {
		t.Errorf("Next3 %d exp: %d", actualT, zeroValueT)
	}
	//Next4
	if actualT, hasValue = iter.Next(); hasValue {
		t.Error("Next4 hasValue true")
	}
	if actualT != zeroValueT {
		t.Errorf("Next4 %d exp: %d", actualT, zeroValueT)
	}

	// SameValue
	iter = NewConverterIterator(NewSliceIterator(slice), fn)
	if actualT = iter.SameValue(); actualT != valueT {
		t.Errorf("SameValue %d exp: %d", actualT, valueT)
	}
	// SameValue2
	if actualT = iter.SameValue(); actualT != valueT {
		t.Errorf("SameValue2 %d exp: %d", actualT, valueT)
	}

	// Cancel
	iter = NewConverterIterator(NewSliceIterator(slice), fn)
	delegator := iter.(*Delegator[int])
	converterItertor := delegator.Delegate.(*ConverterIterator[string, int])
	converterItertor.err = iterErr
	if err = iter.Cancel(); err != iterErr {
		t.Errorf("Cancel1 err: '%v' exp: '%v'", err, iterErr)
	}
	if err = iter.Cancel(); err != iterErr {
		t.Errorf("Cancel2 err: '%v' exp: '%v'", err, iterErr)
	}
	iter = NewConverterIterator(NewSliceIterator(slice), func(key string, isCancel bool) (value int, err error) {
		return
	})
	if err = iter.Cancel(); err != nil {
		t.Errorf("Cancel3 err: '%v''", err)
	}
}

func TestNewConverterIterator(t *testing.T) {
	valueK := "value"
	messageFnNil := "fn cannot be nil"
	messageKeytIteratorNil := "keyIterator cannot be nil"

	var slice = []string{valueK}
	var err error
	fn := func(key string, isCancel bool) (value int, err error) { return }

	NewConverterIterator(NewSliceIterator(slice), fn)

	func() {
		defer func() {
			if v := recover(); v != nil {
				var ok bool
				if err, ok = v.(error); !ok {
					err = perrors.Errorf("panic-value not error; %T '%[1]v'", v)
				}
			}
		}()
		NewConverterIterator[string, int](nil, nil)
	}()
	if err == nil || !strings.Contains(err.Error(), messageFnNil) {
		t.Errorf("NewConverterIterator incorrect panic: '%v' exp %q", err, messageFnNil)
	}

	func() {
		defer func() {
			if v := recover(); v != nil {
				var ok bool
				if err, ok = v.(error); !ok {
					err = perrors.Errorf("panic-value not error; %T '%[1]v'", v)
				}
			}
		}()
		NewConverterIterator(nil, fn)
	}()
	if err == nil || !strings.Contains(err.Error(), messageKeytIteratorNil) {
		t.Errorf("NewConverterIterator incorrect key panic: '%v' exp %q", err, messageKeytIteratorNil)
	}
}

func TestInitConverterIterator(t *testing.T) {
	keyIterator := NewEmptyIterator[string]()
	var iterp ConverterIterator[string, int]
	fn := func(key string, isCancel bool) (value int, err error) { return }

	InitConverterIterator(&iterp, keyIterator, fn)
}
