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
	if s.dataWait.IsActive.Load() || !s.dataWait.IsActive.CompareAndSwap(false, true) {
		return // not first invocation
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

	// one-off initialization of zero-out
	if s.outQ.InQ.MaxRetainSize.Load() == 0 {
		var zeroOut pslib.ZeroOut
		if !prlib.HasReference(t) {
			zeroOut = pslib.NoZeroOut
		}
		s.outQ.InQ.ZeroOut.Store(zeroOut)
	}

	// get size and lowAlloc
	// - effective size must be positive
	//	- minimum defaultSize
	var isLowAlloc bool
	var sizeofT = int(unsafe.Sizeof(t))
	if size < 1 {
		var typeIsError = reflect.TypeFor[T]().Name() == errorTypeName
		if typeIsError {

			// special default handling of type T error
			size = minElements
			isLowAlloc = true
		} else {

			// default size is number of elements in 4 KiB rounded down
			// or minimum 10
			size = max(targetSliceByteSize4KiB/sizeofT, minElements)
		}
	} else {

		// specific size setting small causes low alloc
		isLowAlloc = size <= minElements && size*sizeofT < targetSliceByteSize4KiB
	}
	s.outQ.InQ.Size.Store(size)
	s.outQ.InQ.IsLowAlloc.Store(isLowAlloc)

	// once size established, determine effective allocation size in bytes
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

// postOutput relinquishes outQ lock and
// initializes eventual update of DataWaitCh and EmptyCh
func (s *AwaitableSlice[T]) postOutput() {
	s.outQ.lock.Unlock()
	s.updateWait() // postGet
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
