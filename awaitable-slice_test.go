/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"
	"io"
	"slices"
	"sync"
	"testing"

	"github.com/haraldrudell/parl/pslices/pslib"
)

func TestAwaitableSlice(t *testing.T) {
	const (
		value1, value2, value3 = 1, 2, 3
		size                   = 25
	)

	var (
		actuals []int
	)

	// Get() GetAll() GetSlice() GetSlices() Read() AwaitValue() Seq()
	// Send() SendSlice() SendSlices() SendClone() Write()
	// DataWaitCh() EmptyCh()
	// Close() IsClosed()
	// SetSize() State() String()
	var queue *AwaitableSlice[int]
	var reset = func() {
		queue = &AwaitableSlice[int]{}
	}

	// SetSize should be effective
	reset()
	queue.SetSize(size)
	queue.Send(value1)
	actuals = queue.GetSlice()
	if cap(actuals) != size {
		t.Errorf("FAIL SetSize %d exp %d", cap(actuals), size)
	}
}

func TestAwaitableSliceDataWaitCh(t *testing.T) {
	const (
		value1, value2, value3 = 1, 2, 3
		size                   = 25
	)

	var (
		ch     AwaitableCh
		isOpen bool
	)

	var queue *AwaitableSlice[int]
	var reset = func() {
		queue = &AwaitableSlice[int]{}
	}

	// DataWaitCh
	reset()
	// DataWaitCh on creation should return non-nil, open channel
	ch = queue.DataWaitCh()
	if ch == nil {
		t.Error("FAIL DataWaitCh nil")
	}
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if !isOpen {
		t.Error("FAIL DataWaitCh ch not open")
	}
	// hasData true should close the returned channel
	queue.Send(value1)
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if isOpen {
		t.Error("FAIL DataWaitCh hasData ch not closed")
	}
}

func TestAwaitableSliceCloseCh(t *testing.T) {
	const (
		value1, value2, value3 = 1, 2, 3
		size                   = 25
	)

	var (
		ch               AwaitableCh
		hasValue, isOpen bool
		actual           int
	)

	var queue *AwaitableSlice[int]
	var reset = func() {
		queue = &AwaitableSlice[int]{}
	}

	// CloseCh on creation should return non-nil open channel
	reset()
	ch = queue.CloseCh()
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if !isOpen {
		t.Error("FAIL EndCh empty ch closed")
	}
	// close should close the channel
	queue.Close()
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if isOpen {
		t.Error("FAIL EndCh open after empty Close")
	}

	// CloseCh for hasData true Closed returns open channel
	reset()
	queue.Send(value1)
	queue.Close()
	ch = queue.CloseCh()
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if !isOpen {
		t.Error("FAIL EmptyCh hasData ch closed")
	}
	// hasData to false should close the returned channel
	actual, hasValue = queue.Get()
	_ = actual
	_ = hasValue
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if isOpen {
		t.Error("FAIL EmptyCh empty ch not closed")
	}

	// CloseCh should defer empty detection
	reset()
	ch = queue.CloseCh()
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if !isOpen {
		t.Error("FAIL EmptyCh does not defer empty detection")
	}
	// Close should close the returned channel
	_ = queue.Close()
	select {
	case <-ch:
		isOpen = false
	default:
		isOpen = true
	}
	if isOpen {
		t.Error("FAIL subsequent Close does not close the channel")
	}
}

func TestAwaitableSliceFor(t *testing.T) {
	var value1 = 1
	var expValue1 = []int{value1}

	var a *AwaitableForTester
	var actuals []int

	var slice *AwaitableSlice[int]
	var reset = func() {
		slice = &AwaitableSlice[int]{}
	}

	// Condition active on value appearing and slice closing
	reset()
	// start for loop in other thread
	a = NewAwaitableForTester(slice)
	go a.GoFor()
	<-a.IsReady.Ch()
	// Condition should block not receiving value
	if a.IsValues.IsClosed() {
		t.Fatal("FAIL Condition unexpectedly received value")
	}
	// Condition should not end due to close
	if a.IsClosed.IsClosed() {
		t.Fatal("FAIL Condition IsClosed")
	}
	// Condition should receive appearing values
	slice.Send(value1)
	// Condition should receive value
	<-a.IsValues.Ch()
	actuals = a.Values()
	if !slices.Equal(actuals, expValue1) {
		t.Errorf("FAIL Condition %v exp %v", actuals, expValue1)
	}
	// Condition should not detect a close
	if a.IsClosed.IsClosed() {
		t.Fatal("FAIL Condition IsClosed")
	}
	// Condition should detect occuring close
	slice.Close()
	<-a.IsClosed.Ch()

	// Condition deferred close
	//	- slice has value and is closed on entering Condition
	reset()
	slice.Send(value1)
	slice.Close()
	// start for loop in other thread
	a = NewAwaitableForTester(slice)
	go a.GoFor()
	<-a.IsReady.Ch()
	// Condition should detect close
	<-a.IsClosed.Ch()
	// Condition should have received value
	actuals = a.Values()
	if !slices.Equal(actuals, expValue1) {
		t.Errorf("FAIL Condition %v exp %v", actuals, expValue1)
	}
}

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
		len(slice.outQ.sliceList), slice.outQ.InQ.primary != nil, slice.outQ.InQ.cachedInput != nil,
	)
	slice.outQ.InQ.primary = nil
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

func TestAwaitableSliceGet(t *testing.T) {
	const (
		value1, value2, value3 = 1, 2, 3
		size                   = 25
	)
	var (
		// values [1 2 3]
		values = []int{value1, value2, value3}
	)

	var (
		actual   int
		hasValue bool
	)

	var queue *AwaitableSlice[int]
	var reset = func() {
		queue = &AwaitableSlice[int]{}
	}

	// Get should return one value at a a time
	//	- Send SendSlice should work
	reset()
	queue.Send(value1)
	queue.SendSlice([]int{value2})
	queue.Send(value3)

	// populated Slice: q: [1] slices: [[2] [3]]
	t.Logf("Get: populated Slice: q: %v slices: %v hasData %t",
		queue.outQ.InQ.primary, queue.outQ.InQ.sliceList, queue.outQ.HasDataBits.hasData(),
	)
	for i, v := range values {

		actual, hasValue = queue.Get()
		t.Logf("Get#%d(0…%d) value %d hasValue: %t hasData: %t",
			i, len(values)-1,
			actual, hasValue, queue.outQ.HasDataBits.hasData(),
		)

		if !hasValue {
			t.Errorf("FAIL Get#%d(0…%d) hasValue: %t hasData %t",
				i, len(values)-1, hasValue, queue.outQ.HasDataBits.hasData(),
			)
		}
		if actual != v {
			t.Errorf("FAIL Get#%d(0…%d) %d exp %d  hasData %t",
				i, len(values)-1, actual, v, queue.outQ.HasDataBits.hasData(),
			)
		}
	}
	// Get empty should returns hasValue false
	actual, hasValue = queue.Get()
	_ = actual
	if hasValue {
		t.Error("FAIL Get hasValue true")
	}
}

func TestAwaitableSliceGetSlice(t *testing.T) {
	const (
		value1, value2, value3 = 1, 2, 3
		size                   = 25
	)
	var (
		// values [1 2 3]
		values = []int{value1, value2, value3}
	)

	var (
		actuals []int
	)

	var queue *AwaitableSlice[int]
	var reset = func() {
		queue = &AwaitableSlice[int]{}
	}

	// GetSlice should return one slice at a a time
	// Sednd SendSlice SendSlices should work
	reset()
	queue.Send(value1)
	queue.SendSlice([]int{value2})
	queue.SendSlices([][]int{{value3}})
	t.Logf("GetSlice: populated Slice: q: %v slices: %v hasData %t",
		queue.outQ.InQ.primary, queue.outQ.InQ.sliceList, queue.outQ.HasDataBits.hasData(),
	)
	for i, v := range values {

		actuals = queue.GetSlice()
		t.Logf("GetSlice#%d(0…%d) slice: %v hasData: %t",
			i, len(values)-1,
			actuals, queue.outQ.HasDataBits.hasData(),
		)

		if len(actuals) != 1 {
			t.Fatalf("FAIL Get#%d hasValue false", i)
		}
		if actuals[0] != v {
			t.Errorf("FAIL Get#%d %d exp %d", i, actuals[0], v)
		}
	}
	// Get empty returns nil
	actuals = queue.GetSlice()
	if actuals != nil {
		t.Errorf("FAIL Get actuals not nil: %d%v", len(actuals), actuals)
	}
}

func TestAwaitableSliceGetAll(t *testing.T) {
	//t.Error("loggin on")
	const (
		value1, value2, value3 = 1, 2, 3
		size                   = 25
	)
	var (
		// values [1 2 3]
		values = []int{value1, value2, value3}
	)

	var (
		actuals []int
	)

	var queue *AwaitableSlice[int]
	var reset = func() {
		queue = &AwaitableSlice[int]{}
	}

	// GetAll should return all values in a single slice
	reset()
	queue.Send(value1)
	queue.SendSlice([]int{value2})
	queue.Send(value3)
	// populated Slice: q: [1] slices: [[2] [3]]
	t.Logf("GetAll: populated Slice: q: %v slices: %v hasData %t",
		queue.outQ.InQ.primary, queue.outQ.InQ.sliceList, queue.outQ.HasDataBits.hasData(),
	)

	actuals = queue.GetAll()
	t.Logf("GetAll actual: %v", actuals)

	if !slices.Equal(actuals, values) {
		t.Errorf("FAIL GetAll %v exp %v", actuals, values)
	}
	actuals = queue.GetAll()
	if actuals != nil {
		t.Errorf("FAIL GetAll empty not nil: %d%v", len(actuals), actuals)
	}
}

func TestAwaitableSliceGetSlices(t *testing.T) {
	//t.Error("loggin on")
	const (
		value1, value2, value3 = 1, 2, 3
	)
	var (
		// expValueSlice [ [1] [2 3] ]
		expValueSlice = [][]int{{value1}, {value2, value3}}
	)

	var (
		actuals [][]int
	)

	var queue *AwaitableSlice[int]
	var reset = func() {
		queue = &AwaitableSlice[int]{}
	}

	// GetSlices should return all values in value-slice and sliceList
	reset()
	queue.Send(value1)
	queue.SendSlice([]int{value2})
	queue.Send(value3)

	// populated Slice: q: [1] slices: [[2] [3]]
	t.Logf("GetSlices: populated Slice: q: %v slices: %v hasData %t",
		queue.outQ.InQ.primary, queue.outQ.InQ.sliceList, queue.outQ.HasDataBits.hasData(),
	)

	actuals = queue.GetSlices()
	t.Logf("GetSlices actuals: %v", actuals)

	if !compareSliceLists(actuals, expValueSlice) {
		t.Errorf("FAIL value-slice %v exp %v", actuals, expValueSlice)
	}
	actuals = queue.GetSlices()
	if actuals != nil {
		t.Errorf("FAIL GetSlices empty not nil: %d%v", len(actuals), actuals)
	}
}

func TestAwaitableSliceRead(t *testing.T) {
	//t.Error("loggin on")
	const (
		size      = 25
		bufSize   = 2
		expNRead1 = 2
		expNRead2 = 1
	)
	var (
		value1, value2, value3 = 1, 2, 3
		// expBufRead1 *int [1 2]
		expBufRead1 = []*int{&value1, &value2}
		// expBufRead2 *int [1 2]
		expBufRead2 = []*int{&value3}
	)

	var (
		// 2-element buffer
		buffer    = make([]*int, bufSize)
		n         int
		err       error
		actual    []*int
		actualEOF bool
	)

	var queue *AwaitableSlice[*int]
	var reset = func() {
		queue = &AwaitableSlice[*int]{}
	}

	t.Logf("pointers: 1:0x%x 2:0x%x 3:0x%x",
		&value1, &value2, &value3,
	)

	// GetAll should return all values in a single slice
	reset()
	queue.Send(&value1)
	queue.SendSlice([]*int{&value2})
	queue.Send(&value3)
	// populated Slice: q: [1] slices: [[2] [3]]
	t.Logf("Read: populated Slice: q: %v slices: %v hasData %t",
		queue.outQ.InQ.primary, queue.outQ.InQ.sliceList, queue.outQ.HasDataBits.hasData(),
	)

	// Read should return first two pointers
	n, err = queue.Read(buffer)
	t.Logf("Read n: %d err: %v", n, err)

	if err != nil {
		t.Errorf("FAIL Read err “%s”", err)
	}
	if n != expNRead1 {
		t.Errorf("FAIL Read#1 n %d exp %d", n, expNRead1)
	}
	actual = buffer[:n]
	if !slices.Equal(actual, expBufRead1) {
		t.Errorf("FAIL Read#1 buf %v exp %v", actual, expBufRead1)
	}

	// close enyters drain phase
	err = queue.Close()
	if err != nil {
		t.Errorf("FAIL Close err: “%s”", err)
	}

	// Read should return one more pointer
	n, err = queue.Read(buffer)
	t.Logf("Read n: %d err: %v", n, err)

	actualEOF = errors.Is(err, io.EOF)
	if !actualEOF {
		if err == nil {
			t.Error("FAIL Read#2 missing error")
		} else {
			t.Errorf("FAIL Read#2 bad error: “%s”", err)
		}
	}
	if n != expNRead2 {
		t.Errorf("FAIL Read#2 n %d exp %d", n, expNRead2)
	}
	actual = buffer[:n]
	if !slices.Equal(actual, expBufRead2) {
		t.Errorf("FAIL Read#2 buf %v exp %v", actual, expBufRead2)
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
	for value := range a.slice.Seq {
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

// compareSliceLists returns true if two sliceLists are equal
func compareSliceLists[T comparable](a, b [][]T) (isEqual bool) {
	var aNil, bNil = a == nil, b == nil
	if aNil != bNil {
		return
	}
	if len(a) != len(b) {
		return
	}
	for i := range len(a) {
		if !slices.Equal(a[i], b[i]) {
			return
		}
	}
	isEqual = true
	return
}

func TestAwaitableSliceAlloc(t *testing.T) {
	//t.Error("Logging on")
	var (
		state AwaitableSliceState
		err   error
	)

	var intSlice AwaitableSlice[int]

	intSlice.Send(1)
	intSlice.Get()
	state, _ = intSlice.State()
	// intSlice: {
	// Size:512 MaxRetainSize:512 HasData:0
	// IsDataWaitActive:false IsDataWaitClosed:false
	// IsCloseInvoked:false IsClosed:false
	// Head:{Length:0 Capacity:512} CachedOutput:{Length:0 Capacity:0}
	// OutList:{Length:0 Capacity:512} OutQ:[]
	// Primary:{Length:0 Capacity:512} CachedInput:{Length:0 Capacity:0}
	// InList:{Length:0 Capacity:512} InQ:[]
	// HasInput:false HasList:true ZeroOut:NoZeroOut
	// }
	t.Logf("intSlice: %+v", state)

	// int-queue should have outQ head 4 KiB
	if state.Head.Capacity != sliceListSize {
		t.Errorf("int queue head capacity %d exp %d",
			state.Head.Capacity, sliceListSize,
		)
	}

	// int-queue should have InQ cachedInput pre-alloccated
	if state.CachedInput.Capacity == 0 {
		t.Error("int queue cachedInput not pre-alloc")
	}

	// int-queue should be NoZeroOut
	if state.ZeroOut != pslib.NoZeroOut {
		t.Errorf("int queue bad zero-out: %s exp %s",
			state.ZeroOut, pslib.NoZeroOut,
		)
	}

	var errSlice AwaitableSlice[error]

	// error-queue state should…
	errSlice.Send(err)
	errSlice.Get()
	state, _ = errSlice.State()
	// errSlice: {
	// Size:10 MaxRetainSize:100
	// HasData:0 IsDataWaitActive:false IsDataWaitClosed:false
	// IsCloseInvoked:false IsClosed:false
	// Head:{Length:0 Capacity:10} CachedOutput:{Length:0 Capacity:0}
	// OutList:{Length:0 Capacity:512} OutQ:[]
	// Primary:{Length:0 Capacity:10} CachedInput:{Length:0 Capacity:10}
	// InList:{Length:0 Capacity:512} InQ:[]
	// HasInput:true HasList:true ZeroOut:DoZeroOut
	// }
	t.Logf("errSlice: %+v", state)

	// error-queue should have outQ head 10 errors
	if state.Head.Capacity != minElements {
		t.Errorf("error queue head capacity %d exp %d",
			state.Head.Capacity, minElements,
		)
	}

	// error-queue should have InQ cachedInput nil
	if state.CachedInput.Capacity != 0 {
		t.Error("error queue cachedInput pre-alloc")
	}

	// error-queue should be DoZeroOut
	if state.ZeroOut != pslib.DoZeroOut {
		t.Errorf("error queue bad zero-out: %s exp %s",
			state.ZeroOut, pslib.DoZeroOut,
		)
	}
}
