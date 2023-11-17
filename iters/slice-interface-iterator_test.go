/*
© 2023-present Harald Rudell <haraldrudell@proton.me> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved
*/

package iters

import (
	"testing"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

type siiInterface interface {
	siiMethod()
}

type siiType struct{}

func (t *siiType) siiMethod() {}

var _ siiInterface = &siiType{}

type siiBadType struct{}

// cannot use &siiBadType{} (value of type *siiBadType)
// as siiInterface value in variable declaration:
// *siiBadType does not implement siiInterface (missing method siiMethod)
// var _ siiInterface = &siiBadType{}

// tests Init Cond Next Same Cancel
func TestSliceInterfaceIterator(t *testing.T) {
	var sliceOfConcreteType = make([]siiType, 1)

	var iterator, iter Iterator[siiInterface]
	var value, zeroValue siiInterface
	var hasValue, condition bool

	// Same should return first value
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	value, hasValue = iterator.Same()
	if !hasValue {
		t.Error("Same hasValue false")
	}
	if value != &sliceOfConcreteType[0] {
		t.Error("Same value bad")
	}

	// Same-Same should return first value
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	value, hasValue = iterator.Same()
	_ = value
	_ = hasValue
	value, hasValue = iterator.Same()
	if !hasValue {
		t.Error("Same hasValue false")
	}
	if value != &sliceOfConcreteType[0] {
		t.Error("Same value bad")
	}

	// Cancel-Same should return no value
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	iterator.Cancel()
	value, hasValue = iterator.Same()
	if hasValue {
		t.Error("Cancel-Same hasValue true")
	}
	if value != zeroValue {
		t.Error("Cancel-Same value not zero-value")
	}

	// Cancel-Next should return no value
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	iterator.Cancel()
	value, hasValue = iterator.Next()
	if hasValue {
		t.Error("Cancel-Same hasValue true")
	}
	if value != zeroValue {
		t.Error("Cancel-Same value not zero-value")
	}

	// Cancel-Cond should return no value
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	iterator.Cancel()
	condition = iterator.Cond(&value)
	if condition {
		t.Error("Cancel-Cond condition true")
	}
	if value != zeroValue {
		t.Error("Cancel-Cond value not zero-value")
	}

	// Next should return first value
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	value, hasValue = iterator.Next()
	if !hasValue {
		t.Error("Nex hasValue false")
	}
	if value != &sliceOfConcreteType[0] {
		t.Error("Next value bad")
	}

	// Next-Next should return no value
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	value, hasValue = iterator.Next()
	_ = value
	_ = hasValue
	value, hasValue = iterator.Next()
	if hasValue {
		t.Error("Cancel-Same hasValue true")
	}
	if value != zeroValue {
		t.Error("Cancel-Same value not zero-value")
	}

	// Init should return zero-value and iterator
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	value, iter = iterator.Init()
	if value != zeroValue {
		t.Error("Init value not zero-value")
	}
	if iter != iterator {
		t.Error("Init iterator bad")
	}

	// Init-Cond should return first value and true
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	value, iter = iterator.Init()
	_ = value
	_ = iter
	condition = iterator.Cond(&value)
	if value != &sliceOfConcreteType[0] {
		t.Errorf("Init-Cond bad value: 0x%x exp 0x%x",
			cyclebreaker.Uintptr(value),
			cyclebreaker.Uintptr(&sliceOfConcreteType[0]),
		)
	}
	if !condition {
		t.Error("Init-Cond condition false")
	}

	// Init-Cond-Cond should return no value
	iterator = NewSliceInterfaceIterator[siiInterface](sliceOfConcreteType)
	value, iter = iterator.Init()
	_ = value
	_ = iter
	condition = iterator.Cond(&value)
	_ = condition
	condition = iterator.Cond(&value)
	if value != zeroValue {
		t.Error("Init-Cond-Cond not zero-value")
	}
	if condition {
		t.Error("Init-Cond-Cond condition true")
	}

}

// tests interface I that does not implement E
func TestSliceInterfaceIteratorBad(t *testing.T) {

	isPanic, err := invokeNewSliceInterfaceIterator()
	if !isPanic {
		t.Error("NewSliceInterfaceIterator bad type no panic")
	}
	if err == nil {
		t.Error("NewSliceInterfaceIterator bad type no error")
	}
}

// invokes NewSliceInterfaceIterator recovering and returning an expected panic
func invokeNewSliceInterfaceIterator() (isPanic bool, err error) {
	defer cyclebreaker.RecoverErr(func() cyclebreaker.DA { return cyclebreaker.A() }, &err, &isPanic)

	var sliceOfBadType = make([]siiBadType, 1)
	NewSliceInterfaceIterator[siiInterface](sliceOfBadType)
	return
}
