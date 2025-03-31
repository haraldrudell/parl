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

	"github.com/haraldrudell/parl/perrors"
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
	// hasDataBits is bit-field indicating whether there is data behind locks
	//	- two bits are combined into an atomic providing data intergrity
	//	- — hasDataBit is set when the overall AwaitableSlice has data
	//	- — set behind queueLock
	//	- — cleared behind outputLock using atomic operations
	//	- queueLockHasDataBit is set when there is data behind queueLock
	//	- — written behind queueLock
	//	- this design means outputLock does not have to access
	//		queueLock to determine if there is data behind queueLock
	//	- because queueLock is not synchronized with outputLock,
	//		there must be combined atomic or lock to ensure integrity
	//	-— eg. an ongoing outputLock may overwrite queue bit after
	//		a parallel queueLock sets it to true
	//	- both bits are set to true by simple write in postGet behind queueLock
	//		when data was just added to the queue. For as long as queueLock is held,
	//		state cannot change
	//	- queueLockHasDataBit is cleared while holding both locks when
	//		all queueLock data is moved to outputLock
	//	- later during in the same lock configuration, if it is concluded that
	//		outputLock will be emptied, hasDataBit is cleared, too
	//	- hasDataBit may be cleared during subsequent Get that empties outputLock
	//		but only if queueLockHasDataBit has not been set again ensured by
	//		CompareAndSwap
	//	- value can be read at any time without holding any locks
	hasDataBits hasData
	// inQ contains lock and queues for input methods
	//	- separate input and output locks reduces contention
	inQ inputQueue[T]
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
	if !s.hasDataBits.hasData() {
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
	if s.inQ.maxRetainSize.Load() == 0 {
		var zeroOut pslib.ZeroOut
		if !prlib.HasReference(t) {
			zeroOut = pslib.NoZeroOut
		}
		s.outQ.ZeroOut.Store(zeroOut)
		s.inQ.ZeroOut.Store(zeroOut)
	}

	// get size
	// - effective size must be positive
	//	- minimum defaultSize
	if size < 1 {
		var typeIsError = reflect.TypeFor[T]().Name() == "error"
		if typeIsError {
			// TODO 250327 error no pre-alloc
			size = minElements
		} else {
			// size is number of elements in 4 KiB rounded down
			// or minimum 10
			size = max(targetSliceByteSize4KiB/int(unsafe.Sizeof(t)), minElements)
		}
	}
	s.inQ.size.Store(size)

	// set retain-size to the larger of size and 100
	//	- slices less or equal to this capacity are retained to reduce allocations
	//	- slices larger than this are discarded to avoid temporary memory leaks
	s.inQ.maxRetainSize.Store(max(size, maxForPrealloc))
}

// State returns internal state for debug purposes
//   - holds both locks: not for frequent use
func (s *AwaitableSlice[T]) State() (state AwaitableSliceState) {
	defer s.outQ.lock.Lock().Unlock()
	defer s.inQ.lock.Lock().Unlock()

	state = AwaitableSliceState{
		Size:             s.inQ.size.Load(),
		MaxRetainSize:    s.inQ.maxRetainSize.Load(),
		HasData:          uint32(s.hasDataBits.bits.Load()),
		IsDataWaitActive: s.dataWait.IsActive.Load(),
		IsCloseInvoked:   s.isCloseInvoked.Load(),
		Primary: Metrics{
			Length:   len(s.inQ.primary),
			Capacity: cap(s.inQ.primary),
		},
		CachedInput: Metrics{
			Length:   len(s.inQ.cachedInput),
			Capacity: cap(s.inQ.cachedInput),
		},
		InList: Metrics{
			Length:   len(s.inQ.list.sliceList),
			Capacity: cap(s.inQ.list.sliceList),
		},
		Head: Metrics{
			Length:   len(s.outQ.head),
			Capacity: cap(s.outQ.head),
		},
		CachedOutput: Metrics{
			Length:   len(s.outQ.cachedOutput),
			Capacity: cap(s.outQ.cachedOutput),
		},
		OutList: Metrics{
			Length:   len(s.inQ.list.sliceList),
			Capacity: cap(s.inQ.list.sliceList),
		},
		ZeroOut: s.outQ.ZeroOut.Load(),
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
			var slicep = &s.inQ.list.sliceList[i]
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

	return
}

// “awaitableSlice:error_state:empty_0x14000090000”
//   - error is the name of type T
//   - state can be: uninit data empty drain closed
//   - 0x… is unique identifier: the slice’s memory address
//   - String does not access any lock
func (s *AwaitableSlice[T]) String() (s2 string) {
	var state string
	if s.inQ.size.Load() == 0 {
		state = "uninit"
	} else if !s.isCloseInvoked.Load() {
		if s.hasDataBits.hasData() {
			state = "data"
		} else {
			state = "empty"
		}
	} else if s.isEmpty.IsClosed() {
		state = "closed"
	} else {
		state = "drain"
	}
	var typeName = reflect.TypeFor[T]()
	s2 = fmt.Sprintf("awaitableSlice:%s_state:%s_0x%x",
		typeName, state, Uintptr(s),
	)
	return
}

// enterInputCritical enters output critical section
//   - cannot be invoked while in input critical section
//   - may be invoked while in output critical section
func (s *AwaitableSlice[T]) enterInputCritical() (s2 *AwaitableSlice[T]) {
	s2 = s
	if s.inQ.maxRetainSize.Load() == 0 {
		s.SetSize(0)
	}
	s.inQ.lock.Lock()
	return
}

// enterOutputCritical enters output critical section
//   - cannot be invoked while in input critical section or output critical section
func (s *AwaitableSlice[T]) enterOutputCritical() (s2 *AwaitableSlice[T]) {
	s2 = s
	if s.inQ.maxRetainSize.Load() == 0 {
		s.SetSize(0)
	}
	s.outQ.lock.Lock()
	return
}

// transferToOutQ moves all data to outQ.
// Upon return, all data is with outQ and hasData is up-to-date
//   - action getValue: seeking single value. If slice is longer, it should be saved to outQ head
//   - action getSlice: seeking entire slice
//   - action getNothing: returned slice is nil.
//     outQ may not be empty.
//     used by [AwaitableSlice.GetAll] and [AwaitableSlice.GetSlices] to aggregate all data in outQ.
//   - action getMax: seeking maxCount number of items
//     If slice is longer, it should be saved to s.output.
//     Items should be fetched both from slice and s.outputs
//   - maxCount: optional number of items for action getMax
//   - slice: any non-empty slice from queueLock.
//     slice is not stored in outputLock.
//     Any additional data in queue is moved to outQ head
//   - —
//   - on invocation:
//   - — output is empty unless action is getNothing
//   - — outputLock is held
//   - — hasData is true
//   - to reduce queueLock contention, transferToOutQ:
//   - — transfers all values to outQ and
//   - — transfers pre-allocated slices to inQ
func (s *AwaitableSlice[T]) transferToOutQ(action getAction, maxCount ...int) (slice []T, slices [][]T) {

	// only access queueLock if there is data behind it
	if s.hasDataBits.isInQEmpty() {
		// no data behind queueLock
		if action != getNothing {
			return // entire queue empty return
		}

		// for GetAll GetSlices set hasData false
		//      - if it is still zero, also reset hasDataBit using CompareAndSwap
		s.hasDataBits.setOutputLockEmpty()
		return // only data in outQ return
	}

	slice = s.emptyInQ(action)

	// now update hasData
	//	- hasData is currently true

	switch action {
	case getValue:
		// if fetching single value and more than one value in that slice,
		// not end of data: hasData should remain true
		if len(slice) > 1 || !s.outQ.isEmptyList() {
			return // have more than one value: hasData true
		}
	case getNothing:
		// getAll and getSlices will empty entire queue: hasData false

	case getSlice:
		// slice will be consumed entirely
		if !s.outQ.isEmptyList() {
			return // more slices exist: hasData true
		}
		//	- there are no more slices, so hasData is now false
	case getMax:

		// looking for specific number of items: n
		var n int
		if len(maxCount) > 0 {
			n = maxCount[0]
		}

		var headLength, _ = s.outQ.getHeadMetrics()
		// if n goes negative, hasData should be true
		//	- return keeps hasData true
		//	- if s.output, s.outputs contains more than n: hasData true
		//	- if s.output, s.outputs contains less or equal to n: hasData false
		if n -= headLength; n < 0 {
			return // more items than n: hasData true
		} else if n -= s.outQ.getListElementCount(); n < 0 {
			return // more items than n: hasData true
		}
	default:
		panic(perrors.ErrorfPF("Bad action: %d", action))
	}
	// hasData should be false

	// there is not enough data for the queue to not become empty
	// update hasDataBits while holding inQ lock
	s.hasDataBits.setOutputLockEmpty()

	return // hasData now false return
}

// emptyInQ acquires inQ lock and transfer all elements to outQ
//   - outQ lock must be held
func (s *AwaitableSlice[T]) emptyInQ(action getAction) (slice []T) {

	// prealloc outside inQ lock
	//	- because inQ has data, inQ primary is allocated and non-empty

	var headHasLength bool
	if action == getNothing {
		var length, _ = s.outQ.getHeadMetrics()
		headHasLength = length > 0
	}

	// ensure outQ sliceList is present if required
	if _, c := s.outQ.getListMetrics(); c == 0 {
		// if inQ has slice, outQ likely need one too
		//	- length cannot be checked outside lock
		var needSliceList = s.inQ.HasList.Load()
		if action == getNothing && !needSliceList && headHasLength {
			needSliceList = true
		}
		if needSliceList {
			s.outQ.swapList(s.inQ.makeSliceList())
		}
	}

	// futureSliceList is prealloc for inQ SliceList
	//	- check atomic inQ status if required
	var futureSliceList [][]T
	if !s.inQ.HasList.Load() {
		futureSliceList = s.inQ.makeSliceList()
	}

	// if outQ head is empty, it will become inQ primary
	// if outQ head is non-empty, futurePrimary will become inQ primary
	var futurePrimary = s.preAlloc()
	defer s.inQ.lock.Lock().Unlock()

	// because both locks are held, hasDataBits can be written directly
	//	- clear inQHasDataBit since inQ will be emptied now
	//	- keep hasDataBit set for now
	s.hasDataBits.resetToHasDataBit() // sliceFromQueue
	s.transferCached(futureSliceList)

	// three tasks while holding queueLock:
	//	- find what slices to return
	//	- transfer all other slices to outputLock
	//	- update hasData

	// output and outputLock is empty

	var setPrimary bool

	// try to get data to slice
	switch action {
	case getValue, getSlice:
		slice = s.inQ.swapPrimary()
		setPrimary = true

	case getNothing, getMax:

		// transfer primary to outQ head or
		// outQ sliceList
		var primary = s.inQ.swapPrimary()
		if headHasLength {
			s.outQ.enqueueInList(primary)
			setPrimary = true

		} else {
			//	- instead swap empty allocated outQ head
			//		with any non-empty inQ primary
			// get the other
			var head, _ = s.outQ.swapHead()
			// put both
			s.inQ.swapPrimary(head)
			s.outQ.swapHead(primary)
		}

	default:
		panic(perrors.ErrorfPF("Bad action: %d", action))
	}

	// possibly transfer pre-made head to inQ primary
	if setPrimary {
		if futurePrimary == nil {
			// transfer output to inQ lock
			futurePrimary, _ = s.outQ.swapHead()
		}
		s.inQ.swapPrimary(futurePrimary)
	}

	// transfer any remaining slices
	if !s.inQ.isEmptyList() {
		var slices = s.inQ.getListSlice()
		s.outQ.enqueueInList(slices...)
		// empty and zero-out s.slices
		s.inQ.clearList()
	}

	return // inQ is now empty
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
		if !s.hasDataBits.hasData() {
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
	if s.hasDataBits.hasData() == s.dataWait.Cyclic.IsClosed() {
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
	var hasData = s.hasDataBits.hasData()

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

// appendToValueSlice appends values to valueSlice imited by max
//   - valueSlice was determine to not have enough capacity
func (s *AwaitableSlice[T]) appendToValueSlice(valueSlice, values *[]T) (didChange bool) {
	var v = *valueSlice
	var vals = *values
	var nv0 = len(v)

	// if capacity too large, only copy
	if cap(v) >= maxAppendValueSliceCapacity {
		// n is number of elements available for copy
		var n = cap(v) - nv0
		if n == 0 {
			return // no additional elements
		}
		didChange = true
		n = min(n, len(vals))
		// extend v
		v = v[:nv0+n]
		copy(v[nv0:], vals[:n])
		*valueSlice = v
		if n == len(vals) {
			*values = nil
			return // entire values via copy
		}
		*values = vals[n:]
		return // some of values copied
	}
	// valueSlice can be extended
	didChange = true

	// n is initial allowed append length,
	// depends on current length
	var n = maxAppendValueSliceCapacity - len(v)
	// adjust for vals length
	var n0 = min(n, len(vals))
	v = append(v, vals[:n0]...)
	if len(vals) == n0 {
		*valueSlice = v
		*values = nil
		return // all appended return
	}
	vals = vals[n0:]
	// there are more in vals

	// copy any extra from excessive realloc
	n = cap(v) - len(v)
	if n == 0 {
		return
	}
	nv0 = len(v)
	n0 = min(n, len(vals))
	v = v[:nv0+n0]
	copy(v[nv0:], vals[:n0])
	*valueSlice = v
	if n0 == len(vals) {
		*values = nil
		return
	}
	*values = vals[n0:]

	return
}

// append to sliceList

// preAlloc onlyCached
//const onlyCachedTrue = true

// preAlloc ensures that output0 and cachedOutput are allocated
// to configured value-slice size
//   - must hold outputLock
//   - onlyCached missing or false: both output and cachedOutput are
//     ensured to be allocate and reusable
//   - onlyCached true: by former GetAll: only cachedOutput is ensured
//     to be allocated and reusable
//   - purpose is to reduce allocations while holding queueLock
func (s *AwaitableSlice[T]) preAlloc() (futurePrimary []T) {

	// when created inQ does not have q or sliceList allocated
	//	- Send needs q
	//	- SendClone SendSlice needs sliceList if q non-empty
	//	- SendClone SendSlice discard q if sliceList empty

	var length, capacity = s.outQ.getHeadMetrics()

	if length > 0 {
		futurePrimary = s.inQ.makeValueSlice()
	} else {

		// output0 first pre-allocation
		//	- ensure output0 is a slice of good capacity
		//	- may be transferred to queueLock
		//	- avoids allocation while holding queueLock
		// should output0 be allocated?
		var makeOutput = capacity == 0
		if !makeOutput {
			// reuse for capacities defaultSize–maxRetainSize
			makeOutput = capacity < minElements || capacity > s.inQ.maxRetainSize.Load()
		}
		if makeOutput {
			s.outQ.swapHead(s.inQ.makeValueSlice())
		}
	}

	// cachedOutput second pre-allocation
	//	- possibly have ready for transfer
	//	- configured size may be large so only for defaultSize
	if !s.outQ.hasCachedOutput() && s.inQ.size.Load() == minElements {
		s.outQ.swapCachedOutput(s.inQ.makeValueSlice())
	}

	return
}

// transferCached transfers cachedOutput from
// outputLock to queueLock if possible
//   - invoked while holding outputLock queueLock
func (s *AwaitableSlice[T]) transferCached(sliceList [][]T) {

	// transfer any pre-allocated sliceList
	if sliceList != nil && !s.inQ.HasList.Load() {
		s.inQ.swapList(sliceList)
		s.inQ.HasList.Store(true)
	}

	// transfer cachedOutput to cachedInput
	if !s.inQ.HasInput.Load() && s.outQ.hasCachedOutput() {
		var slice = s.outQ.swapCachedOutput()
		s.inQ.setCachedInput(slice)
		s.outQ.swapCachedOutput(nil)
	}
}

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
			s.hasDataBits.setOutputLockEmpty() //postGet
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
	s.hasDataBits.setAllBits() // postSend
	s.inQ.lock.Unlock()
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
	minElements = 10
	// max appended size is 16 MiB
	maxAppendValueSliceCapacity = 16 * 1024 * 1024
	// scavenging: max size for slice preallocation
	maxForPrealloc = 100
	// IntSize is 4 on 32-bit and 8 on 64-bit architectures as constant
	IntSize = 4 * (1 + math.MaxUint>>63)
	// sliceListSize is allocation-size for slice lists
	sliceListSize = targetSliceByteSize4KiB / IntSize
)

const (
	// sliceFromQueue is invoked to get a single value
	//	- [AwaitableSlice.Get]
	getValue getAction = iota
	// sliceFromQueue is invoked to get a non-empty slice of values
	//	- [AwaitableSlice.GetSlice]
	getSlice
	// sliceFromQueue is invoked to move all elements to outQ
	//	- [AwaitableSlice.GetSlices] [AwaitableSlice.GetAll]
	getNothing
	// sliceFromQueue is invoked to get maxCount elements
	//	- [AwaitableSlice.Read]
	getMax
)

// action for sliceFromQueue
//   - getValue getSlice getNothing getMax
type getAction uint8
