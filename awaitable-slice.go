/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"slices"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/pslices"
)

const (
	// default allocation size for new slices if Size is < 1
	defaultSize = 10
	// scavenging: max size for slice preallocation
	maxForPrealloc = 100
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
//   - [AwaitableSlice.EndCh] returns a channel that closes on slice empty,
//     configurable to provide close-like behavior
//   - [AwaitableSlice.SetSize] allows for setting initial slice capacity
//   - AwaitableSlice benefits:
//   - — trouble-free data-sink: non-blocking-unbound-send, non-deadlocking, panic-free and error-free
//   - — initialization-free, awaitable, thread-less and thread-safe
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
//   - AwaitableSlice deprecates:
//   - — [NBChan] fully-featured unbound channel
//   - — [NBRareChan] low-usage unbound channel
//
// Usage:
//
//	var valueQueue parl.AwaitableSlice[*Value]
//	go func(valueSink parl.ValueSink) {
//	  defer valueSink.EmptyCh()
//	  …
//	  valueSink.Send(value)
//	  …
//	}(&valueQueue)
//	endCh := valueQueue.EmptyCh(parl.CloseAwait)
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
	// allocation size for new slices, effective if > 0
	//	- 10 or larger value from SetSize
	size Atomic64[int]
	// maxRetainSize is the longest  slice that will be reused
	//	- avoids temporary memory leaks
	maxRetainSize Atomic64[int]
	// queueLock makes queue and slices thread-safe
	//	- queueLock also makes Send SendSlice critical sections
	queueLock sync.Mutex
	// queue is a locally made slice for individual values
	//	- behind queueLock
	//	- not a slice-away slie
	queue []T
	// slices contains slices of values transferred by SendSlice and
	// possible subsequent locally made slices of values
	//	- slice-away slice, behind queueLock
	slices, slices0 [][]T
	// isLocalSlice is true if the last slice of slices is locally made
	//	- only valid when slices non-empty
	//	- behind queueLock
	isLocalSlice bool
	// indicates at all times whether the queue is empty
	//	- allows for updateDataWait to be invoked without any locks held
	//	- written behind queueLock
	hasData atomic.Bool
	// a pre-allocated slice for queue
	//	- behind queueLock
	//	- allocated by Get Get1 GetAll prior to acquiring queueLock
	cachedInput []T
	// outputLock makes output thread-safe
	//	- outputLock also makes Get1 Get critical sections
	outputLock sync.Mutex
	// output is a slice being sliced away from
	//	- behind outputLock, slice-away slice
	output, output0 []T
	// outputs contains entire-slice values
	//	- behind outputLock, slice-away slice
	outputs, outputs0 [][]T
	// a pre-allocated slice for queue
	//	- behind outputLock
	//	- allocated by Get Get1 GetAll prior to acquiring queueLock
	cachedOutput []T
	// lazy DataWaitCh
	dataWait LazyCyclic
	// lazy emptyWait
	emptyWait LazyCyclic
}

// Send enqueues a single value. Thread-safe
func (s *AwaitableSlice[T]) Send(value T) {
	defer s.postSend()
	s.queueLock.Lock()

	// add to queue if no slices
	if len(s.slices) == 0 {
		if s.queue != nil {
			s.queue = append(s.queue, value)
			// create s.queue
		} else if s.cachedInput != nil {
			// use cachedInput allocated under outputLock
			s.queue = append(s.cachedInput, value)
			s.cachedInput = nil
		} else {
			s.queue = s.make(value)
		}
		return // value in queue return
	}

	// add to slices
	//	- if last slice not locally created, append to new slice
	if s.isLocalSlice {
		// append to ending local slice
		var index = len(s.slices) - 1
		s.slices[index] = append(s.slices[index], value)
		return
	}

	// append local slice
	var q []T
	if s.cachedInput != nil {
		q = append(s.cachedInput, value)
		s.cachedInput = nil
	} else {
		q = s.make(value)
	}
	pslices.SliceAwayAppend1(&s.slices, &s.slices0, q)
	s.isLocalSlice = true
}

// SendSlice provides values by transferring ownership of a slice to the queue
//   - SendSlice may reduce allocations and increase performance by handling multiple values
//   - Thread-safe
func (s *AwaitableSlice[T]) SendSlice(values []T) {
	// ignore empty slice
	if len(values) == 0 {
		return
	}
	defer s.postSend()
	s.queueLock.Lock()

	// append to slices
	s.slices = append(s.slices, values)
	s.isLocalSlice = false
}

// SendClone provides a value-slice without transferring ownership of a slice to the queue
//   - allocation
//   - Thread-safe
func (s *AwaitableSlice[T]) SendClone(values []T) {
	// ignore empty slice
	if len(values) == 0 {
		return
	}
	s.SendSlice(slices.Clone(values))
}

// DataWaitCh returns a channel that closes once values becomes available
//   - Thread-safe
func (s *AwaitableSlice[T]) DataWaitCh() (ch AwaitableCh) {
	// this may initialize the cyclic awaitable
	ch = s.dataWait.Cyclic.Ch()

	// if previously invoked, no need for initialization
	if s.dataWait.IsActive.Load() {
		return // not first invocation
	}

	// establish proper state
	//	- data wait ch now in use
	if !s.dataWait.IsActive.CompareAndSwap(false, true) {
		return
	}

	// set initial state
	s.updateWait()

	return
}

// [AwaitableSlice.EmptyCh] initialize: this invocation
// will wait for close-like state, do not activate EmptyCh awaitable
const CloseAwaiter = false

// EmptyCh returns an awaitable channel that closes on queue being or
// becoming empty
//   - doNotInitialize missing: enable closing of ch which will happen as soon
//     as the slice is empty, possibly prior to return
//   - doNotInitialize CloseAwaiter: obtain the channel but do not enable it closing.
//     A subsequent invocation with doNotInitialize missing will enable its closing thus
//     act as a deferred Close function
//   - thread-safe
func (s *AwaitableSlice[T]) EmptyCh(doNotInitialize ...bool) (ch AwaitableCh) {
	// this may initialize the cyclic awaitable
	ch = s.emptyWait.Cyclic.Ch()

	// if previously invoked, no need for initialization
	if len(doNotInitialize) > 0 || s.emptyWait.IsActive.Load() {
		return // not first invocation
	}

	// establish proper state
	//	- data wait ch now in use
	if !s.emptyWait.IsActive.CompareAndSwap(false, true) {
		return
	}

	// set initial state
	if !s.hasData.Load() {
		s.emptyWait.Cyclic.Close()
	}

	return
}

// Get returns one value if the queue is not empty
//   - hasValue true: value is valid
//   - hasValue false: the queue is empty
//   - Get may attain allocation-free receive or allocation-free operation
//   - — a slice is not returned
//   - — an internal slice may be reused reducing allocations
//   - thread-safe
func (s *AwaitableSlice[T]) Get() (value T, hasValue bool) {
	if !s.hasData.Load() {
		return
	}
	defer s.postGet()
	s.outputLock.Lock()

	// if output empty, transfer outputs[0] to output
	if len(s.output) == 0 && len(s.outputs) > 0 {
		// possibly save output to cachedOutput
		if c := cap(s.output); c == defaultSize && s.cachedOutput == nil {
			s.cachedOutput = s.output0
		}
		// write new s.output
		var so = s.outputs[0]
		s.output = so
		s.output0 = so
		s.outputs[0] = nil
		s.outputs = s.outputs[1:]
	}

	// try output
	if hasValue = len(s.output) > 0; hasValue {
		value = s.output[0]
		var zeroValue T
		s.output[0] = zeroValue
		s.output = s.output[1:]
		return // got value from s.output
	}

	// transfer from queueLock
	var slice = s.sliceFromQueue(isOne)
	if hasValue = len(slice) > 0; !hasValue {
		return // no value available
	}

	// store slice as output and fetch value
	s.output0 = slice
	value = slice[0]
	var zeroValue T
	slice[0] = zeroValue
	s.output = slice[1:]

	return
}

// GetSlice returns a slice of values from the queue
//   - values non-nil: a non-empty slice at a time, not necessarily all data.
//     values is never non-nil and empty
//   - — Send-GetSlice: each GetSlice empties the queue
//   - — SendMany-GetSlice: each GetSlice receives one SendMany slice
//   - values nil: the queue is empty
//   - GetSlice may increase performance by slice-at-a-time operation, however,
//     slices need to be allocated:
//   - — Send-GetSlice requires internal slice allocation
//   - — SendMany-GetSlice requires sender to allocate slices
//   - — Send-Get1 may reduce allocations
//   - thread-safe
func (s *AwaitableSlice[T]) GetSlice() (values []T) {
	if !s.hasData.Load() {
		return
	}
	defer s.postGet()
	s.outputLock.Lock()

	// try output
	if len(s.output) > 0 {
		values = s.output
		s.output0 = nil
		s.output = nil
		return
	}

	// try s.outputs
	var so = s.outputs
	if len(so) > 0 {
		values = so[0]
		so[0] = nil
		s.outputs = so[1:]
		return
	}

	// transfer from queueLock
	//	- values may be nil
	values = s.sliceFromQueue(isSlice)

	return
}

// GetAll returns a single slice of all unread values in the queue
//   - values nil: the queue is empty
//   - thread-safe
func (s *AwaitableSlice[T]) GetAll() (values []T) {
	if !s.hasData.Load() {
		return
	}
	defer s.postGet()
	s.outputLock.Lock()

	// aggregate outputLock data
	// output is a copy of s.output since preAlloc may destroy it
	var size = len(s.output)
	for _, o := range s.outputs {
		size += len(o)
	}

	// aggregate queueLock data
	s.preAlloc(onlyCachedTrue)
	defer s.queueLock.Unlock()
	s.queueLock.Lock()
	s.transferCached()

	// aggregate queueLock data
	size += len(s.queue)
	for _, s := range s.slices {
		size += len(s)
	}
	// size is now length of the returned slice

	// no data
	if size == 0 {
		return // no data return
	}

	// will return all data so queue will be empty
	//	- because size is not zero, hasData is changing
	//	- written while holding queueLock
	s.hasData.Store(false)

	// attempt allocation-free single slice return
	if values = s.singleSlice(size); len(values) > 0 {
		return // single slice
	}

	// create aggregate slice
	values = make([]T, 0, size)
	if len(s.output) > 0 {
		values = append(values, s.output...)
		pslices.SetLength(&s.output, 0)
	}
	for _, s := range s.outputs {
		values = append(values, s...)
	}
	pslices.SetLength(&s.outputs, 0)
	if len(s.queue) > 0 {
		values = append(values, s.queue...)
		pslices.SetLength(&s.queue, 0)
	}
	for _, s := range s.slices {
		values = append(values, s...)
	}
	pslices.SetLength(&s.slices, 0)

	return
}

// Init allows for AwaitableSlice to be used in a for clause
//   - returns zero-value for a short variable declaration in
//     a for init statement
//   - thread-safe
//
// Usage:
//
//	var a AwaitableSlice[…] = …
//	for value := a.Init(); a.Condition(&value); {
//	  // process received value
//	}
//	// the AwaitableSlice closed
func (s *AwaitableSlice[T]) Init() (value T) { return }

// Condition allows for AwaitableSlice to be used in a for clause
//   - updates a value variable and returns whether values are present
//   - thread-safe
//
// Usage:
//
//	var a AwaitableSlice[…] = …
//	for value := a.Init(); a.Condition(&value); {
//	  // process received value
//	}
//	// the AwaitableSlice closed
func (s *AwaitableSlice[T]) Condition(valuep *T) (hasValue bool) {
	var endCh AwaitableCh
	for {

		// try obtaining value
		if s.hasData.Load() {
			var v T
			if v, hasValue = s.Get(); hasValue {
				*valuep = v
				return // value obtained: *valuep valid, hasValue true
			}
			continue
		}
		// hasData is false
		//	- wait until the slice has data or
		//	- the slice closes

		// atomic-performance check for channel end
		if s.emptyWait.IsActive.Load() {
			// channel is out of items and closed
			return // closed: hasValue false, *valuep unchanged
		}

		// await data or close
		if endCh == nil {
			// get endCh without initializing close mechanic
			endCh = s.EmptyCh(CloseAwaiter)
		}
		select {

		// await data, possibly initializing dataWait
		case <-s.DataWaitCh():

			// await close and end of data
		case <-endCh:
			return // closed: hasValue false, *valuep unchanged
		}
	}
}

// SetSize set initial allocation size of slices. Thread-safe
func (s *AwaitableSlice[T]) SetSize(size int) {
	var maxSize int
	if size < 1 {
		size = defaultSize
	} else if size > maxForPrealloc {
		maxSize = size
	} else {
		maxSize = maxForPrealloc
	}
	s.size.Store(size)
	s.maxRetainSize.Store(maxSize)
}

// make returns a new slice of length 0 and configured capacity
//   - value, if present, is added to the new slice
func (s *AwaitableSlice[T]) make(value ...T) (newSlice []T) {

	// ensure size sizeMax are initialized
	var size = s.size.Load()
	if size == 0 {
		s.SetSize(0)
		size = s.size.Load()
	}

	//create slice optionally with value
	newSlice = make([]T, len(value), size)
	if len(value) > 0 {
		newSlice[0] = value[0]
	}

	return
}

const (
	isOne   = false
	isSlice = true
)

// sliceFromQueue fetches slices from queue to output
//   - getSlice true: seeking entire slice
//   - getSlice false: seeking single value
//   - invoked when output empty
func (s *AwaitableSlice[T]) sliceFromQueue(getSlice bool) (slice []T) {
	//prealloc outside queueLock
	s.preAlloc()
	s.queueLock.Lock()
	defer s.queueLock.Unlock()
	s.transferCached()

	// three tasks while holding queueLock:
	//	- find what slice to return
	//	- transfer all other slices to outputLock
	//	- update hasData

	// retrieve queue if non-empty
	if len(s.queue) > 0 {
		slice = s.queue
		s.queue = nil
	}
	// possibly transfer pre-made output0 to queueLock
	if s.queue == nil {
		// transfer output0 to queueLock
		s.queue = s.output0
		s.output0 = nil
	}

	// if slice empty, try first of slices
	if len(slice) == 0 && len(s.slices) > 0 {
		slice = s.slices[0]
		s.slices[0] = nil
		s.slices = s.slices[1:]
	}

	// transfer any remaining slices
	if len(s.slices) > 0 {
		pslices.SliceAwayAppend(&s.outputs, &s.outputs0, s.slices)
		// empty and zero-out s.slices
		pslices.SetLength(&s.slices, 0)
		s.slices = s.slices0[:0]
	}

	// hasData must be updated while holding queueLock
	//	- it is currently true
	if len(s.outputs) > 0 {
		return // slices in outputs mean data still available: no change

		// if fetching single value and more than one value in that slice,
		// not end of data
	} else if !getSlice && len(slice) > 1 {
		return // no, single-value fetch and more than one value
	}

	// the queue is empty: update hasData while holding queueLock
	s.hasData.Store(false)

	return
}

// setData updates dataWaitCh if DataWaitCh was invoked
//   - eventually consistent
//   - atomized hasData observation with dataWait emptyWait update
//   - shielded atomic performance
func (s *AwaitableSlice[T]) updateWait() {

	// dataWait closes on data available, then re-opens
	//	- hasData true: dataWait should be closed
	//	- emptyWait closes on empty and remains closed
	//	- hasData observed false, emptyWait should be closed

	// is dataWait or emptyWait in use?
	//	- both only go once from false to true
	var dataWait = s.dataWait.IsActive.Load()

	if dataWait {
		// atomic check based on dataWait
		if s.hasData.Load() == s.dataWait.Cyclic.IsClosed() {
			return // atomically state was ok
		}
		// atomic check based on emptyWait
	} else if !s.emptyWait.IsActive.Load() {
		return // neither is active
		// close emptyCh if empty
	} else if s.hasData.Load() || s.emptyWait.Cyclic.IsClosed() {
		return // atomically state was ok
	}

	// alter state using the lock of dataWait
	//	- even if dataWait is not active
	//	- atomizes hasData observation with Open/Close operation
	s.dataWait.Lock.Lock()
	defer s.dataWait.Lock.Unlock()

	// hasData inside lock
	var hasData = s.hasData.Load()

	// if dataWait active:
	if dataWait {
		// check against dataWait state
		if hasData == s.dataWait.Cyclic.IsClosed() {
			return // no change
		} else if hasData {
			// hasData true: close dataWait
			s.dataWait.Cyclic.Close()
			// emptyWait does not re-open
		} else {
			// hasData false: open dataWait
			s.dataWait.Cyclic.Open()
			// hasData false: trigger emptyWait
			if s.emptyWait.IsActive.Load() {
				s.emptyWait.Cyclic.Close()
			}
			return
		}
	}
	// emptyWait is active, dataWait inactive

	// if not empty, no action
	if hasData {
		return
	}
	// no data: trigger emptyWait
	s.emptyWait.Cyclic.Close()
}

// ensureSize ensures that size and maxRetainSize are initialized
//   - size: the configured allocation-size of a new queue slice
func (s *AwaitableSlice[T]) ensureSize() (size int) {
	// ensure size sizeMax are initialized
	if size = s.size.Load(); size == 0 {
		s.SetSize(0)
		size = s.size.Load()
	}
	return
}

// preAlloc onlyCached
const onlyCachedTrue = true

// preAlloc ensures that output0 and cachedOutput are allocated
// to configured size
//   - must hold outputLock
func (s *AwaitableSlice[T]) preAlloc(onlyCached ...bool) {

	var size = s.ensureSize()
	if len(onlyCached) == 0 || !onlyCached[0] {

		// output0 first pre-allocation
		//	- ensure output0 is a slice of good capacity
		//	- may be transferred to queueLock
		//	- avoids allocation while holding queueLock
		// should output0 be allocated?
		var makeOutput = s.output0 == nil
		if !makeOutput {
			// check capacity of existing output
			var c = cap(s.output0)
			// reuse for capcities defaultSize–maxRetainSize
			makeOutput = c < defaultSize || c > s.maxRetainSize.Load()
		}
		if makeOutput {
			var so = s.make()
			s.output0 = so
			s.output = so
		}
	}

	// cachedOutput second pre-allocation
	//	- possibly have ready for transfer
	//	- configured size may be large so only for defaultSize
	if s.cachedOutput == nil && size == defaultSize {
		s.cachedOutput = s.make()
	}
}

// transferCached transfers cachedOutput from
// outputLock to queueLock if possible
//   - invoked while holding outputLock queueLock
func (s *AwaitableSlice[T]) transferCached() {

	// transfer cachedOutput to queueLock
	if s.cachedInput == nil && s.cachedOutput != nil {
		s.cachedInput = s.cachedOutput
		s.cachedOutput = nil
	}
}

// postGet relinquishes outputLock and
// initializes eventual update of DataWaitCh and EmptyCh
//   - aggregates deferred actions to reduce latency
//   - invoked while holding outputLock
func (s *AwaitableSlice[T]) postGet() {
	s.outputLock.Unlock()
	s.updateWait()
}

// singleSlice fetches values if contained in a single slice
//   - reduces slice allocations by using an existing slice
//   - invoked while holding outputLock queueLock
func (s *AwaitableSlice[T]) singleSlice(size int) (values []T) {

	// only output
	if size == len(s.output) {
		values = s.output
		s.output = nil
		s.output0 = nil
		return // got values
	} else if len(s.output) > 0 {
		return // is aggregate
	}

	// only outputs[0]
	if len(s.outputs) == 1 && size == len(s.outputs[0]) {
		values = s.outputs[0]
		s.outputs[0] = nil
		s.outputs = s.outputs[1:]
		return // got values
	} else if len(s.outputs) > 0 {
		return // is aggregate
	}

	// only queue
	if len(s.queue) == size {
		values = s.queue
		s.queue = nil
	}
	// possibly transfer pre-made output0 to queueLock
	if s.queue == nil {
		// transfer output0 to queueLock
		s.queue = s.output0
		s.output0 = nil
	}
	if len(s.queue) > 0 || len(s.queue) == size {
		return // got values or is aggregate
	}

	// only s.slices[0]
	if len(s.slices) == 1 {
		values = s.slices[0]
		s.slices = s.slices[1:]
	}

	return // got values or is aggregate
}

// postSend set hasData true, relinquishes queueuLock and
// initializes eventual update of DataWaitCh and EmptyCh
//   - aggregates deferred actions to reduce latency
//   - invoked while holding queueLock
func (s *AwaitableSlice[T]) postSend() {
	s.hasData.Store(true)
	s.queueLock.Unlock()
	s.updateWait()
}
