/*
© 2022-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"strings"
	"testing"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
	"golang.org/x/exp/slices"
)

func TestSliceIterator(t *testing.T) {
	slice := []string{"one", "two"}
	var zeroValue string

	var err error
	var iter SliceIterator[string]
	var value string
	var hasValue bool

	InitSliceIterator(&iter, slices.Clone(slice))

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
}

func TestNewSliceIterator(t *testing.T) {
	slice := []string{}
	var zeroValue string

	var iter Iterator[string]
	var value string
	var hasValue bool

	iter = NewSliceIterator(slices.Clone(slice))

	if value, hasValue = iter.Same(); hasValue {
		t.Error("Same hasValue true")
	}
	if value != zeroValue {
		t.Error("Same hasValue not zeroValue")
	}
}

func TestInitSliceIterator(t *testing.T) {
	slice := []string{"one"}
	messageIterpNil := "cannot be nil"

	var iter SliceIterator[string]
	var value string
	var hasValue bool
	var iterpNil *SliceIterator[string]
	var err error

	iter = SliceIterator[string]{}
	InitSliceIterator(&iter, slices.Clone(slice))

	if value, hasValue = iter.Next(IsSame); !hasValue {
		t.Error("Same hasValue false")
	}
	if value != slice[0] {
		t.Errorf("Same value %q exp %q", value, slice[0])
	}

	cyclebreaker.RecoverInvocationPanic(func() {
		InitSliceIterator(iterpNil, slice)
	}, &err)
	if err == nil || !strings.Contains(err.Error(), messageIterpNil) {
		t.Errorf("InitSliceIterator incorrect panic: '%v' exp %q", err, messageIterpNil)
	}
}
