/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"slices"
	"testing"
)

func TestAwaitableSlice(t *testing.T) {
	var value1, value2, value3 = 1, 2, 3
	var values = []int{value1, value2, value3}
	var size = 25

	var actual int
	var actuals []int
	var hasValue, isOpen bool
	var ch AwaitableCh

	// DataWaitCh Get Get1 Send SendSlice SetSize
	var slice *AwaitableSlice[int]
	var reset = func() {
		slice = &AwaitableSlice[int]{}
	}

	// Get1 returns one value at a a time
	//	- Send SendSlice
	reset()
	slice.Send(value1)
	slice.SendSlice([]int{value2})
	slice.Send(value3)
	// populated Slice: q: [1] slices: [[2] [3]]
	t.Logf("populated Slice: q: %v slices: %v", slice.queue, slice.slices)
	for i, v := range values {
		actual, hasValue = slice.Get1()
		if !hasValue {
			t.Errorf("Get1#%d hasValue false", i)
		}
		if actual != v {
			t.Errorf("Get1#%d %d exp %d", i, actual, v)
		}
	}
	// Get1 empty returns hasValue false
	actual, hasValue = slice.Get1()
	_ = actual
	if hasValue {
		t.Error("Get1 hasValue true")
	}

	// Get returns a slice at a a time
	reset()
	slice.Send(value1)
	slice.SendSlice([]int{value2})
	slice.Send(value3)
	for i, v := range values {
		actuals = slice.Get()
		if len(actuals) != 1 {
			t.Fatalf("Get#%d hasValue false", i)
		}
		if actuals[0] != v {
			t.Errorf("Get#%d %d exp %d", i, actuals[0], v)
		}
	}
	// Get empty returns nil
	actuals = slice.Get()
	if actuals != nil {
		t.Errorf("Get actuals not nil: %d%v", len(actuals), actuals)
	}

	// GetAll returns single slice
	reset()
	slice.Send(value1)
	slice.SendSlice([]int{value2})
	slice.Send(value3)
	actuals = slice.GetAll()
	if !slices.Equal(actuals, values) {
		t.Errorf("GetAll %v exp %v", actuals, values)
	}
	actuals = slice.GetAll()
	if actuals != nil {
		t.Errorf("GetAll empty not nil: %d%v", len(actuals), actuals)
	}

	// SetSize should be effective
	reset()
	slice.SetSize(size)
	slice.Send(value1)
	actuals = slice.Get()
	if cap(actuals) != size {
		t.Errorf("SetSize %d exp %d", cap(actuals), size)
	}

	// DataWaitCh starts open
	reset()
	// DataWaitCh empty should return non-nil open channel
	ch = slice.DataWaitCh()
	if ch == nil {
		t.Error("DataWaitCh nil")
	}
	isOpen = true
	select {
	case <-ch:
		isOpen = false
	default:
	}
	if !isOpen {
		t.Error("DataWaitCh ch not open")
	}
	// DataWaitCh hasData should close
	slice.Send(value1)
	isOpen = true
	select {
	case <-ch:
		isOpen = false
	default:
	}
	if isOpen {
		t.Error("DataWaitCh hasData ch not closed")
	}

	// EndCh for empty returns Closed
	reset()
	ch = slice.EmptyCh()
	isOpen = true
	select {
	case <-ch:
		isOpen = false
	default:
	}
	if isOpen {
		t.Error("EndCh empty ch not closed")
	}

	// EndCh for hasData returns Open
	reset()
	slice.Send(value1)
	ch = slice.EmptyCh()
	isOpen = true
	select {
	case <-ch:
		isOpen = false
	default:
	}
	if !isOpen {
		t.Error("EndCh hasData ch closed")
	}
	// EndCh closes on out-of-data
	actual, hasValue = slice.Get1()
	_ = actual
	_ = hasValue
	isOpen = true
	select {
	case <-ch:
		isOpen = false
	default:
	}
	if isOpen {
		t.Error("EndCh empty ch not closed")
	}
}
