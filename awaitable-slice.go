/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"sync/atomic"
	"unsafe"

	"github.com/haraldrudell/parl/preflect/prlib"
	"github.com/haraldrudell/parl/pslices/pslib"
)

// AwaitableSlice is a queue as thread-safe awaitable unbound slice of element value T or slices of value T
//   - [AwaitableSlice.Send] [AwaitableSlice.Get] allows efficient
//     transfer of single values
//   - [AwaitableSlice.SendSlice] [AwaitableSlice.GetSlice] allows efficient
//     transfer of slices where:
//     a sender relinquish slice ownership by invoking SendSlice and
//     a receiving thread gains slice ownership by invoking GetSlice
//   - lower performing [AwaitableSlice.SendClone]
//   - [AwaitableSlice.DataWaitCh] returns a channel that closes once data is available
//     making the queue awaitable
//   - [AwaitableSlice.CloseCh] returns a channel that closes on slice empty,
//     configurable to provide close-like behavior
//   - [AwaitableSlice.SetSize] allows for setting initial slice capacity
//   - AwaitableSlice benefits:
//   - — #1 many-to-many thread-synchronization mechanic
//   - — #2 trouble-free, closable value-sink: non-blocking unbound send, near-non-deadlocking, panic-free and error-free object
//   - — #3 initialization-free, awaitable and observable, thread-less and thread-safe
//   - — features channel-based wait usable with Go select and default:
//     a consumer may wait for many events or poll for value or close
//   - — unbound with tunable low-allocation
//   - — contention-separation between Send SendSlice SendClone and Get GetSlice
//   - — high-throughput multiple-value operation using SendSlice GetSlice
//   - — slice size logic avoids large-slice memory leaks
//   - — zero-out of unused slice elements avoids temporary memory leaks
//   - — although the slice can transfer values almost allocation free or
//     multiple values at a time,
//     the wait mechanic requires pointer allocation 10 ns,
//     channel make 21 ns, channel close 9 ns as well as
//     CAS operation 8/21 ns
//   - compared to Go channel:
//   - — #1: has no errors or panics, blocking send or thread requirements
//   - — #2: many-to-many: many threads can await slice data or close event, a thread can await many slices or events
//   - — #3: initialization-free, observable with unbound size
//   - — unbound, non-blocking-send that is error and panic free
//   - — happens-before with each received value or detection of value avaliable or close:
//     similar to unbuffered channel guarantees while being buffered
//   - — closable by any thread without race condition
//   - — observable, idempotent, panic-free and error-free close
//     while also able to transmit values
//   - — closing channel one-to-many mechanic for awaiting data and close.
//     Data synchronization is [sync.Mutex] and the queue is slice.
//     All is shielded by atomic performance
//   - — for high parallelism, AwaitableSlice sustains predominately atomic performance while
//     channel has 100× deteriorating unshielded lock performance as of go1.22.3
//
// Usage:
//
//	var valueQueue parl.AwaitableSlice[*Value]
//	go func(valueSink parl.ValueSink) {
//	  defer valueSink.Close()
//	  …
//	  valueSink.Send(value)
//	  …
//	}(&valueQueue)
//	endCh := valueQueue.EmptyCh()
//	for {
//	  select {
//	  case <-valueQueue.DataWaitCh():
//	    for value := valueQueue.Init(); valueQueue.Condition(&value); {
//	      doSomething(value)
//	    }
//	  case <-endCh:
//	    break
//	}
//	// the slice closed
//	…
//	// to reduce blocking, at most 100 at a time
//	for i, hasValue := 0, true; i < 100 && hasValue; i++ {
//	  var value *Value
//	  if value, hasValue = valueQueue.Get(); hasValue {
//	    doSomething(value)
type AwaitableSlice[T any] struct {
	// outQ contains lock and queues for output methods
	//	- separate input and output locks reduces contention
	outQ outputQueue[T]
	// dataWait provides a channel that remains open while the queue is empty and
	// closes upon data becoming available
	//	- dataWait enables consumers to block until data becomes available:
	//		<-q.DataWaitCh()
	//	- the dataWait channel is returned by [AwaitableSlice.DataWaitCh]
	//	- each DataWaitCh invocation may return a different channel value
	//	- dataWait is lazily initialized: if DataWaitCh or [AwaitableSlice.AwaitValue]
	//		are never invoked, dataWait is never initialized
	dataWait LazyCyclic
	// isCloseInvoked indicates in drain phase or closed
	//	- set to true by [AwaitableSlice.Close] invocation
	//	- once read to end, the queue enters closed state with
	//		isEmpty triggered and [AwaitableSlice.IsClosed] returning true
	//	- state changes to drain to closed states are single-transition irreversible
	isCloseInvoked atomic.Bool
	// isEmpty indicates the queue has been closed and read to end: closed state
	//	- closed state means the queue was or became empty after
	//		[AwaitableSlice.Close] was invoked
	//	- state changes to drain and closed states are single-transition irreversible
	//	- isEmpty only creates a channel if [AwaitableSlice.EmptyCh] or
	//		[AwaitableSlice.AwaitValue] were invoked
	isEmpty Awaitable
}

// AwaitableSlice is [IterableSource]
var _ IterableSource[int] = &AwaitableSlice[int]{}

// AwaitableSlice is [ClosableAllSource]
var _ ClosableAllSource[int] = &AwaitableSlice[int]{}

// AwaitableSlice is [Sink]
var _ Sink[int] = &AwaitableSlice[int]{}

// AwaitableSlice is [io.ReadWriteCloser]
var _ io.ReadWriteCloser = &AwaitableSlice[byte]{}

// DataWaitCh returns a channel that is open on empty and closes once values are available
//   - each DataWaitCh invocation may return a different channel value
//   - Thread-safe
func (s *AwaitableSlice[T]) DataWaitCh() (ch AwaitableCh) {

	// invoking Ch may initialize the cyclic awaitable
	ch = s.dataWait.Cyclic.Ch() // DataWaitCh

	// if previously invoked, no need for initialization
	if s.dataWait.IsActive.Load() {
		return // not first invocation
	}

	// establish proper state
	//	- data wait ch now in use
	if !s.dataWait.IsActive.CompareAndSwap(false, true) {
		return // another thread initialized dataWait
	}

	// set initial state of dataWait
	s.updateWait() // initial upon DataWaitCh

	return
}

// CloseCh returns an awaitable closing channel that closes once
// [AwaitableSlice.Close] invoked and the queue being or
// becoming empty
//   - CloseCh always returns the same channel value
//   - thread-safe
func (s *AwaitableSlice[T]) CloseCh() (ch AwaitableCh) { return s.isEmpty.Ch() }

// Close closes the queue
//   - err: nil
//   - Send methods are not affected by Close
//   - Close makes AwaitableSlice implement [io.Closer]
func (s *AwaitableSlice[T]) Close() (err error) {

	// establish proper state
	//	- data wait ch now in use
	if s.isCloseInvoked.Load() || !s.isCloseInvoked.CompareAndSwap(false, true) {
		return // isClose was already triggered return
	}
	// this thread should close the queue

	// set initial isEmpty state
	//	- if queue is already empty, ie. hasData false,
	//		isEmpty should be triggered
	if !s.outQ.HasDataBits.hasData() {
		s.isEmpty.Close()
	}

	return
}

// IsClosed returns true if closable is closed or triggered
//   - isClosed is a single boolean value usable with for or if
//   - IsClosed wraps a 6-line read into a single-value boolean expression
func (s *AwaitableSlice[T]) IsClosed() (isClosed bool) {

	// if emptyCh has not been initialized,
	// the slice is not closed
	if !s.isCloseInvoked.Load() {
		return // isEmpty not initialized return: isClosed false
	}
	// retrieve status of close
	isClosed = s.isEmpty.IsClosed()

	return
}

// SetSize set initial allocation size of slices
//   - size: allocation size in T elements for new value-slices
//   - size < 1: use allocation the larger of 4 KiB or 10 elements
//   - maxRetainSize is set to the larger of size and 100
//   - thread-safe
func (s *AwaitableSlice[T]) SetSize(size int) {

	var t T
	var sizeofT = int(unsafe.Sizeof(t))
	if s.outQ.InQ.MaxRetainSize.Load() == 0 {
		var zeroOut pslib.ZeroOut
		if !prlib.HasReference(t) {
			zeroOut = pslib.NoZeroOut
		}
		s.outQ.InQ.ZeroOut.Store(zeroOut)
	}

	var isLowAlloc bool

	// get size
	// - effective size must be positive
	//	- minimum defaultSize
	if size < 1 {
		var typeIsError = reflect.TypeFor[T]().Name() == errorTypeName
		if typeIsError {
			size = minElements
			isLowAlloc = true
		} else {
			// size is number of elements in 4 KiB rounded down
			// or minimum 10
			size = max(targetSliceByteSize4KiB/sizeofT, minElements)
		}
	} else {
		isLowAlloc = size <= minElements && size*sizeofT < targetSliceByteSize4KiB
	}
	s.outQ.InQ.Size.Store(size)
	s.outQ.InQ.IsLowAlloc.Store(isLowAlloc)

	var sizeBytes = size * sizeofT
	s.outQ.sizeMax4KiB.Store(sizeBytes <= targetSliceByteSize4KiB)

	// set retain-size to the larger of size and 100
	//	- slices less or equal to this capacity are retained to reduce allocations
	//	- slices larger than this are discarded to avoid temporary memory leaks
	s.outQ.InQ.MaxRetainSize.Store(max(size, maxForPrealloc))
}

// Length returns current and historic max length of the queue
//   - length: current number of elements in the queue
//   - maxLength: the highest number of elements that were in the queue
//   - on first invocation of Length, queue begins to track length
//   - using Length has a performance impact
//   - maxLength can only be tracked inside locks
//   - if maxLength is to be obtained, length must be tracked, too
func (s *AwaitableSlice[T]) Length() (length, maxLength int) {

	// ensure length tracking is initialized
	if !s.outQ.InQ.IsLength.Load() {
		s.outQ.initLength()
	}

	length = s.outQ.InQ.Length.Load()
	maxLength = int(s.outQ.InQ.MaxLength.Max1())
	return
}

// State returns internal state for debug purposes
//   - populateValues missing: values are not retrieved
//   - populateValues [ValuesYes]: every queue slice value is provided in values
//   - holds both locks: not for frequent use
func (s *AwaitableSlice[T]) State(populateValues ...ValuesFlag) (state AwaitableSliceState, values AwaitableSliceValues[T]) {
	defer s.outQ.lock.Lock().Unlock()
	defer s.outQ.InQ.lock.Lock().Unlock()

	state = AwaitableSliceState{
		Size:          s.outQ.InQ.Size.Load(),
		MaxRetainSize: s.outQ.InQ.MaxRetainSize.Load(),
		SizeMax4KiB:   s.outQ.sizeMax4KiB.Load(),
		IsLowAlloc:    s.outQ.InQ.IsLowAlloc.Load(),
		ZeroOut:       s.outQ.InQ.ZeroOut.Load(),

		IsDataWaitActive: s.dataWait.IsActive.Load(),
		IsCloseInvoked:   s.isCloseInvoked.Load(),

		HasInput:         s.outQ.InQ.HasInput.Load(),
		HasList:          s.outQ.InQ.HasList.Load(),
		HasData:          uint32(s.outQ.HasDataBits.bits.Load()),
		IsLength:         s.outQ.InQ.IsLength.Load(),
		Length:           s.outQ.InQ.Length.Load(),
		MaxLength:        s.outQ.InQ.MaxLength.Max1(),
		LastPrimaryLarge: s.outQ.lastPrimaryLarge,

		Head: Metrics{
			Length:   len(s.outQ.head),
			Capacity: cap(s.outQ.head),
		},
		CachedOutput: Metrics{
			Length:   len(s.outQ.cachedOutput),
			Capacity: cap(s.outQ.cachedOutput),
		},
		OutList: Metrics{
			Length:   len(s.outQ.InQ.list.sliceList),
			Capacity: cap(s.outQ.InQ.list.sliceList),
		},

		Primary: Metrics{
			Length:   len(s.outQ.InQ.primary),
			Capacity: cap(s.outQ.InQ.primary),
		},
		CachedInput: Metrics{
			Length:   len(s.outQ.InQ.cachedInput),
			Capacity: cap(s.outQ.InQ.cachedInput),
		},
		InList: Metrics{
			Length:   len(s.outQ.InQ.list.sliceList),
			Capacity: cap(s.outQ.InQ.list.sliceList),
		},
	}
	if state.IsDataWaitActive {
		state.IsDataWaitClosed = s.dataWait.Cyclic.IsClosed()
	}
	if state.IsCloseInvoked {
		state.IsClosed = s.isEmpty.IsClosed()
	}

	if x := state.InList.Length; x > 0 {
		state.InQ = make([]Metrics, x)
		for i := range x {
			var slicep = &s.outQ.InQ.list.sliceList[i]
			state.InQ[i].Length = len(*slicep)
			state.InQ[i].Capacity = cap(*slicep)
		}
	}

	if x := state.OutList.Length; x > 0 {
		state.OutQ = make([]Metrics, x)
		for i := range x {
			var slicep = &s.outQ.list.sliceList[i]
			state.OutQ[i].Length = len(*slicep)
			state.OutQ[i].Capacity = cap(*slicep)
		}
	}

	if len(populateValues) == 0 || populateValues[0] != ValuesYes {
		return
	}

	if x := len(s.outQ.head); x > 0 {
		values.Head = make([]T, x)
		copy(values.Head, s.outQ.head)
	}
	if x := len(s.outQ.sliceList); x > 0 {
		values.Outputs = make([][]T, x)
		for i, src := range s.outQ.sliceList {
			var dest = &values.Outputs[i]
			*dest = make([]T, len(src))
			copy(*dest, src)
		}
	}

	if x := len(s.outQ.InQ.primary); x > 0 {
		values.Primary = make([]T, x)
		copy(values.Primary, s.outQ.InQ.primary)
	}
	if x := len(s.outQ.InQ.sliceList); x > 0 {
		values.Inputs = make([][]T, x)
		for i, src := range s.outQ.InQ.sliceList {
			var dest = &values.Inputs[i]
			*dest = make([]T, len(src))
			copy(*dest, src)
		}
	}

	return
}

// “awaitableSlice:int_state:empty_0x14000090000”
//   - int is the name of type T
//   - state can be: uninit data empty drain closed
//   - 0x… is unique identifier: the slice’s memory address
//   - String does not access any lock
func (s *AwaitableSlice[T]) String() (s2 string) {

	// state is the queue’s current state “data”
	var state string
	if s.outQ.InQ.Size.Load() == 0 {
		state = "uninit"
	} else if !s.isCloseInvoked.Load() {
		if s.outQ.HasDataBits.hasData() {
			state = "data"
		} else {
			state = "empty"
		}
	} else if s.isEmpty.IsClosed() {
		state = "closed"
	} else {
		state = "drain"
	}

	var reflectType = reflect.TypeFor[T]()

	s2 = fmt.Sprintf("awaitableSlice:%s_state:%s_0x%x",
		reflectType, state, Uintptr(s),
	)
	return
}

// enterInputCritical enters output critical section
//   - cannot be invoked while in input critical section
//   - may be invoked while in output critical section
func (s *AwaitableSlice[T]) enterInputCritical() (s2 *AwaitableSlice[T]) {
	s2 = s
	if s.outQ.InQ.MaxRetainSize.Load() == 0 {
		s.SetSize(0)
	}
	s.outQ.InQ.lock.Lock()
	return
}

// enterOutputCritical enters output critical section
//   - cannot be invoked while in input critical section or output critical section
func (s *AwaitableSlice[T]) enterOutputCritical() (s2 *AwaitableSlice[T]) {
	s2 = s
	if s.outQ.InQ.MaxRetainSize.Load() == 0 {
		s.SetSize(0)
	}
	s.outQ.lock.Lock()
	return
}

// updateWait updates dataWaitCh if either DataWaitCh or AwaitValue were invoked
//   - eventually consistent and invoked after every Send or Get, ie.
//     at some point after every Send or Get, any active dataWaitCh or EndCh will
//     accurately reflect the latest state
//   - because dataWait and isEmpty update requires multiple separate operations,
//     those are atomized behind dedicated lock dataWait.Lock
//   - shielded atomic performance
//   - dedicated lock separates contention for increased perormance
func (s *AwaitableSlice[T]) updateWait() {

	// dataWait closes on data available, then re-opens
	//	- hasData true: dataWait should be closed
	//	- emptyWait closes on empty and remains closed
	//	- hasData observed false, emptyWait should be closed

	// isEmpty is activated by Close
	//	- because isCloseInvoked and isEmpty are both
	//		single irreversible transitions, deterministic end-state
	//		can be obtained without critical section
	if s.isCloseInvoked.Load() && !s.isEmpty.IsClosed() {
		if !s.outQ.HasDataBits.hasData() {
			// if EmptyCh was invoked, costs channel close.
			// Otherwise, with the 2025 Awaitable: cheap atomics all the way
			s.isEmpty.Close()
		}
	}

	// if false, no action required for dataWait
	if !s.dataWait.IsActive.Load() {
		return // dataWait inactive return
	}

	// use atomics to try to avoid slower lock
	//	- if atomics at any point are able to confirm correct state,
	//		no further action is required
	if s.outQ.HasDataBits.hasData() == s.dataWait.Cyclic.IsClosed() {
		return // dataWait confirmed atomically return
	}
	// dataWait is active and incorrect state observed

	// valdiate state using the lock of dataWait
	//	-	atomizing hasData read with cyclic operation guarantees that
	//		upon completion of the cyclic operation, it correctly reflects a
	//		hasData observation
	//	- without lock, end-state lacks integrity
	defer s.dataWait.Lock.Lock().Unlock()

	// latest hasData value inside lock
	var hasData = s.outQ.HasDataBits.hasData()

	// to close
	if hasData {
		// Close on closed costs the same as isclose check
		//	- because the channel has been retrieved,
		//		close on opened costs channel close
		s.dataWait.Cyclic.Close()
		return
	}

	// to open
	//	- open on opened costs the same as isclose check
	//	- because the channel has been retrieved,
	//		open on closed is alloc expensive
	s.dataWait.Cyclic.Open()
}

// append to sliceList

// preAlloc onlyCached
//const onlyCachedTrue = true

// postOutput relinquishes outQ lock and
// initializes eventual update of DataWaitCh and EmptyCh
//   - gotValue: true if the Get GetSlice GetAll Read operation retrieved a value
//   - checkedQueue: true if queueLock data was checked
//   - —
//   - The reason postOutput has goValue is that if outQ lock was emptied,
//     inQ must be checked to update hasData
//   - this check is carried out on gotValue true, checkedQueue false
//   - postOutput aggregates deferred actions to reduce latency
//   - postOutput is invoked while holding outQ lock
func (s *AwaitableSlice[T]) postOutput(gotValue, checkedQueue *bool) {

	// if a value was retrieved, hasData may need update
	//	- if gotValue is true, hasData was true
	//	- — if both outQ lock and queueLock are now empty
	//	- — then hasData must be set to false
	//	- if queueLock was checked, hasData was already updated
	//	- used by Get GetSlice
	if *gotValue && !*checkedQueue {
		// hasData is true
		// outQ may be empty

		// check if there are any more values in outputLock
		if s.outQ.isEmptyOutput() {
			// if queueLockHasData is false,
			// hasData should be false
			s.outQ.HasDataBits.setOutputLockEmpty() //postGet
		}
	}

	// reliquish outputLock and initiate eventually consistent data/close update
	s.outQ.lock.Unlock()
	s.updateWait() // postGet
}

// postInput sets hasData to true, relinquishes queueLock and
// initializes eventual update of DataWaitCh and EmptyCh
//   - aggregates deferred actions to reduce latency
//   - invoked while holding inQ Lock
func (s *AwaitableSlice[T]) postInput() {
	// overwrite both data bits atomically
	s.outQ.HasDataBits.setAllBits() // postSend
	s.outQ.InQ.lock.Unlock()
	s.updateWait() // postSend
}

// getAwait returns
//   - endCh: closing on Close-empty
//   - dataWait: cyclic with data wait channel
func (s *AwaitableSlice[T]) getAwait() (endCh AwaitableCh, dataWait *CyclicAwaitable) {
	endCh = s.CloseCh()

	// ensure data-wait is initialized
	if !s.dataWait.IsActive.Load() {
		// winner thread does updateWait
		s.DataWaitCh()
	}
	dataWait = &s.dataWait.Cyclic // AwaitValue

	return
}

const (
	// targetSliceByteSize4KiB is the target byte-size for new slices
	//	- used if [AwaitableSlice.SetSize] < 1
	targetSliceByteSize4KiB = 4 * 1024
	// minimum number of element for value-slice
	//	- used as size when type T is error
	//	- when size equal or less, triggers isLowAlloc behavior
	//	- minium size for slice retaining
	minElements = 10
	// max appended size is 16 MiB
	maxAppendValueSliceCapacity = 16 * 1024 * 1024
	// scavenging: max size for slice preallocation
	maxForPrealloc = 100
	// IntSize is 4 on 32-bit and 8 on 64-bit architectures as constant
	IntSize = 4 * (1 + math.MaxUint>>63)
	// sliceListSize is allocation-size for slice lists
	sliceListSize = targetSliceByteSize4KiB / IntSize
	// sliceListSize for low alloc
	lowAllocListSize = 1
	// when allocating sliceList, no minimum size
	noMinSize = 0
	// queue treats T type error special
	errorTypeName = "error"
)
