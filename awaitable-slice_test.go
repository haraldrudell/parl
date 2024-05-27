/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"slices"
	"sync"
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

	// DataWaitCh EmptyCh Get Get1 GetAll Send SendSlice SetSize
	var slice *AwaitableSlice[int]
	var reset = func() {
		slice = &AwaitableSlice[int]{}
	}

	// Get1 should return one value at a a time
	//	- Send SendSlice should work
	reset()
	slice.Send(value1)
	slice.SendSlice([]int{value2})
	slice.Send(value3)
	// populated Slice: q: [1] slices: [[2] [3]]
	t.Logf("populated Slice: q: %v slices: %v", slice.queue, slice.slices)
	for i, v := range values {
		actual, hasValue = slice.Get()
		if !hasValue {
			t.Errorf("Get1#%d hasValue false", i)
		}
		if actual != v {
			t.Errorf("Get1#%d %d exp %d", i, actual, v)
		}
	}
	// Get1 empty should returns hasValue false
	actual, hasValue = slice.Get()
	_ = actual
	if hasValue {
		t.Error("Get1 hasValue true")
	}

	// Get should return one slice at a a time
	reset()
	slice.Send(value1)
	slice.SendSlice([]int{value2})
	slice.Send(value3)
	for i, v := range values {
		actuals = slice.GetSlice()
		if len(actuals) != 1 {
			t.Fatalf("Get#%d hasValue false", i)
		}
		if actuals[0] != v {
			t.Errorf("Get#%d %d exp %d", i, actuals[0], v)
		}
	}
	// Get empty returns nil
	actuals = slice.GetSlice()
	if actuals != nil {
		t.Errorf("Get actuals not nil: %d%v", len(actuals), actuals)
	}

	// GetAll should return all values in a single slice
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
	actuals = slice.GetSlice()
	if cap(actuals) != size {
		t.Errorf("SetSize %d exp %d", cap(actuals), size)
	}

	// DataWaitCh
	reset()
	// DataWaitCh on creation should return non-nil, open channel
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
	// hasData true should close the returned channel
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

	// EndCh on creation should return non-nil closed channel
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

	// EndCh for hasData true returns open channel
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
		t.Error("EmptyCh hasData ch closed")
	}
	// hasData to false should close the returned channel
	actual, hasValue = slice.Get()
	_ = actual
	_ = hasValue
	isOpen = true
	select {
	case <-ch:
		isOpen = false
	default:
	}
	if isOpen {
		t.Error("EmptyCh empty ch not closed")
	}

	// EmptyCh CloseAwaiter should defer empty detection
	reset()
	ch = slice.EmptyCh(CloseAwaiter)
	isOpen = true
	select {
	case <-ch:
		isOpen = false
	default:
	}
	if !isOpen {
		t.Error("EmptyCh CloseAwaiter doe not defer empty detection")
	}
	// EmptyCh without CloseAwaiter should close the returned channel
	_ = slice.EmptyCh()
	isOpen = true
	select {
	case <-ch:
		isOpen = false
	default:
	}
	if isOpen {
		t.Error("subsequent EmptyCh does not close the channel")
	}
}

func TestAwaitableSliceFor(t *testing.T) {
	var value1 = 1
	var expValue1 = []int{value1}

	var actual, zeroValue int
	var a *AwaitableForTester
	var actuals []int

	var slice *AwaitableSlice[int]
	var reset = func() {
		slice = &AwaitableSlice[int]{}
	}

	// Init should return zero-value
	reset()
	actual = slice.Init()
	if actual != zeroValue {
		t.Errorf("Init %d exp %d", actual, zeroValue)
	}

	// Condition active on value appearing and slice closing
	reset()
	// start for loop in other thread
	a = NewAwaitableForTester(slice)
	go a.GoFor()
	<-a.IsReady.Ch()
	// Condition should block not receiving value
	if a.IsValues.IsClosed() {
		t.Fatal("Condition unexpectedly received value")
	}
	// Condition should not end due to close
	if a.IsClosed.IsClosed() {
		t.Fatal("Condition IsClosed")
	}
	// Condition should receive appearing values
	slice.Send(value1)
	// Condition should receive value
	<-a.IsValues.Ch()
	actuals = a.Values()
	if !slices.Equal(actuals, expValue1) {
		t.Errorf("Condition %v exp %v", actuals, expValue1)
	}
	// Condition should not detect a close
	if a.IsClosed.IsClosed() {
		t.Fatal("Condition IsClosed")
	}
	// Condition should detect occuring close
	slice.EmptyCh()
	<-a.IsClosed.Ch()

	// Condition deferred close
	//	- slice has value and is closed on entering Condition
	reset()
	slice.Send(value1)
	slice.EmptyCh()
	// start for loop in other thread
	a = NewAwaitableForTester(slice)
	go a.GoFor()
	<-a.IsReady.Ch()
	// Condition should detect close
	<-a.IsClosed.Ch()
	// Condition should have received value
	actuals = a.Values()
	if !slices.Equal(actuals, expValue1) {
		t.Errorf("Condition %v exp %v", actuals, expValue1)
	}
}

// cover 100% of Send
func TestAwaitableSliceSend(t *testing.T) {
	//t.Error("loggin on")
	var value = 1
	var values = []int{value}

	var slice *AwaitableSlice[int]
	var reset = func() {
		slice = &AwaitableSlice[int]{}
	}

	// use cachedInput in Send
	reset()
	// 2 Get should cover append to s.queue
	slice.Send(value)
	slice.Send(value)
	// Get to empty initializes cachedInput
	slice.GetAll()
	// len(s.slices): 0 s.queue: true s.cachedInput: true
	t.Logf("len(s.slices): %d s.queue: %t s.cachedInput: %t",
		len(slice.slices), slice.queue != nil, slice.cachedInput != nil,
	)
	slice.queue = nil
	// Send should cover using cachedInput
	slice.Send(value)

	// SendSlice and 2 Sends should cover append to local slice
	reset()
	slice.SendSlice(slices.Clone(values))
	slice.Send(value)
	slice.Send(value)

	// use cachedInput for local slice
	reset()
	// Get to empty initializes cachedInput
	slice.Send(value)
	slice.Get()
	// SendSlice then Send should creates a local slice from cachedInput
	slice.SendSlice(values)
	slice.Send(value)
}

// edge cases for Get
func TestAwaitableSliceGet(t *testing.T) {
	var (
		value1, value2 = 1, 2
	)

	var (
		value    int
		hasValue bool
		endCh    AwaitableCh
	)

	var slice *AwaitableSlice[int]
	var reset = func() {
		slice = &AwaitableSlice[int]{}
	}

	// test Get of last item in output in deferred close
	reset()
	slice.Send(value1)
	slice.Send(value2)
	// close stream entering deferred close
	endCh = slice.EmptyCh()
	// Get transfers both items to outputLock
	// Get should return first item
	value, hasValue = slice.Get()
	if !hasValue {
		t.Error("hasValue false")
	}
	if value != value1 {
		t.Errorf("Get %d exp %d", value, value1)
	}
	// stream should not be closed
	select {
	case <-endCh:
		t.Errorf("stream closed")
	default:
	}
	// Get should return last item and close the stream
	value, hasValue = slice.Get()
	if !hasValue {
		t.Error("hasValue false")
	}
	if value != value2 {
		t.Errorf("Get %d exp %d", value, value2)
	}
	// stream should be closed
	select {
	case <-endCh:
	default:
		t.Errorf("stream not closed")
	}
	// Get should retrieve no items
	value, hasValue = slice.Get()
	_ = value
	if hasValue {
		t.Error("hasValue true")
	}
}

// 100% coverage GetSlice
func TestAwaitableSliceGetSlice(t *testing.T) {
	//t.Error("loggin on")
	var value1, value2 = 1, 2
	var value2Slice = []int{value2}

	var actual int
	var hasValue bool
	var actuals []int

	var slice *AwaitableSlice[int]
	var reset = func() {
		slice = &AwaitableSlice[int]{}
	}

	// use cachedInput in Send
	reset()
	// Send Send Get creates non-empty s.output
	slice.Send(value1)
	slice.Send(value2)
	if actual, hasValue = slice.Get(); actual != value1 {
		t.Errorf("Get %d exp %d", actual, value1)
	}
	_ = hasValue
	// GetSlice should return s.output
	if actuals = slice.GetSlice(); !slices.Equal(actuals, value2Slice) {
		t.Errorf("GetSlice %v exp %v", actuals, value2Slice)
	}
}

// 100% coverage GetAll
func TestAwaitableSliceGetAll(t *testing.T) {
	//t.Error("loggin on")
	// value1 1
	var value1, value2, value3 = 1, 2, 3
	// slice of value 1
	var values1, values2, values3 = []int{value1}, []int{value2}, []int{value3}
	var values23 = []int{value2, value3}

	var actual int
	var actuals []int
	var hasValue bool

	var slice *AwaitableSlice[int]
	var reset = func() {
		slice = &AwaitableSlice[int]{}
	}

	// aggregate outputs
	reset()
	// SendSlice SendSlice Get creates non-empty s.outputs
	slice.SendSlice(slices.Clone(values1))
	slice.SendSlice(slices.Clone(values2))
	if actual, hasValue = slice.Get(); actual != value1 {
		t.Errorf("Get %d exp %d", actual, value1)
	}
	_ = hasValue
	// GetAll should return s.outputs
	if actuals = slice.GetAll(); !slices.Equal(actuals, values2) {
		t.Errorf("GetAll %v exp %v", actuals, values2)
	}

	// aggregate output and outputs
	reset()
	// Send Send SendSlice Get creates non-empty s.output s.outputs
	slice.Send(value1)
	slice.Send(value2)
	slice.SendSlice(slices.Clone(values3))
	if actual, hasValue = slice.Get(); actual != value1 {
		t.Errorf("Get %d exp %d", actual, value1)
	}
	_ = hasValue
	// GetAll should aggregate output and outputs
	if actuals = slice.GetAll(); !slices.Equal(actuals, values23) {
		t.Errorf("GetAll %v exp %v", actuals, values23)
	}

	// GetAll only output
	reset()
	// Send Send Get creates non-empty output
	slice.Send(value1)
	slice.Send(value2)
	if actual, hasValue = slice.Get(); actual != value1 {
		t.Errorf("Get %d exp %d", actual, value1)
	}
	_ = hasValue
	// GetAll should return output single-slice
	if actuals = slice.GetAll(); !slices.Equal(actuals, values2) {
		t.Errorf("GetAll %v exp %v", actuals, values2)
	}

	// only s.slices[0]
	reset()
	// SendSlice creates single-slice s.slices
	slice.SendSlice(values1)
	// GetAll should return s.slices[0] single-slice
	if actuals = slice.GetAll(); !slices.Equal(actuals, values1) {
		t.Errorf("GetAll %v exp %v", actuals, values1)
	}
}

type AwaitableForTester struct {
	slice     *AwaitableSlice[int]
	IsReady   Awaitable
	IsClosed  Awaitable
	IsValues  CyclicAwaitable
	valueLock sync.Mutex
	values    []int
}

func NewAwaitableForTester(slice *AwaitableSlice[int]) (a *AwaitableForTester) {
	return &AwaitableForTester{slice: slice}
}

func (a *AwaitableForTester) GoFor() {
	a.IsReady.Close()
	for value := a.slice.Init(); a.slice.Condition(&value); {
		a.addValue(value)
	}
	a.IsClosed.Close()
}

func (a *AwaitableForTester) addValue(value int) {
	a.valueLock.Lock()
	defer a.IsValues.Close()
	defer a.valueLock.Unlock()

	a.values = append(a.values, value)
}

func (a *AwaitableForTester) Values() (values []int) {
	a.valueLock.Lock()
	defer a.valueLock.Unlock()

	values = slices.Clone(a.values)
	return
}
