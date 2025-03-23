/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"
	"slices"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
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
//   - [AwaitableSlice.EmptyCh] returns a channel that closes on slice empty,
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
	// size is allocation size for new slices, set by [AwaitableSlice.SetSize]
	//	- effective slice allocation capacity:
	//	- if size is unset or less than 1: 10: defaultSize
	//	- otherwise, the value provided to [AwaitableSlice.SetSize]
	size Atomic64[int]
	// maxRetainSize is the longest slice capacity that will be reused
	//	- not retaining large slices avoids temporary memory leaks
	//	- reusing slices of reasonable size reduces allocations
	//	- effective value depends on size set by [AwaitableSlice.SetSize]
	//	- if size is unset or less than 100, use maxForPrealloc: 100
	//	- otherwise, use value provided to [AwaitableSlice.SetSize]
	maxRetainSize Atomic64[int]
	// queueLock makes queue and slices thread-safe
	//	- queueLock also makes Send SendSlice critical sections
	//	- high-contention lock
	//	- two locks separates contention of sink and source
	queueLock Mutex
	// queue is a slice of values from Send
	//	- behind queueLock
	//	- not a slice-away slice
	//	- only added to when s.queuen is empty
	//	- queue may be empty while s.queuen is non-empty
	queue []T
	// qSos contains received entire qSos and qSos of values from [AwaitableSlice.Send]
	//	- slice-away slice, behind queueLock
	//	- no value-slice is empty
	qSos, qSos0 [][]T
	// isLocalSlice is true if the last slice of slices is locally made
	//	- only valid when slices non-empty
	//	- behind queueLock
	isLocalSlice bool
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
	//	- the bits can be read at any time without any locks
	hasDataBits Atomic32[hasDataBitField]
	// cachedInput is a pre-allocated value-slice behind queueLock
	//	- used for s.queue or local slice added to s.slices
	//	- copied from cachedOutput while holding both locks
	//	- cachedOutput is allocated by Get Get1 GetAll prior to acquiring queueLock
	cachedInput []T
	// outputLock makes output slices thread-safe
	//	- outputLock also makes Get GetSlice GetAll critical sections
	//	- two locks separates contention of sink and source
	outputLock Mutex
	// output is a slice being sliced away from
	//	- behind outputLock, slice-away slice
	output, output0 []T
	// outputs contains entire-slice values
	//	- behind outputLock, slice-away slice
	outputs, outputs0 [][]T
	// a pre-allocated value-slice tarnsfered to queueLock
	//	- behind outputLock
	//	- allocated by Get Get1 GetAll prior to acquiring queueLock
	cachedOutput []T
	// dataWait provides a channel that remains open while the queue is empty and
	// closes upon data becoming available
	//	- the dataWait channel is returned by [AwaitableSlice.DataWaitCh]
	//	- dataWait enables consumers to block until data becomes available:
	//		<-q.DataWaitCh()
	//	- each DataWaitCh invocation may return a different channel value
	//	- dataWait is lazily initialized: if DataWaitCh or [AwaitableSlice.AwaitValue]
	//		are never invoked, dataWait is never initialized
	dataWait LazyCyclic
	// isCloseInvoked indicates drain phase started by Close
	//	- set to true by [AwaitableSlice.Close] invocation
	//	- once read to end, the queue enters closed state with
	//		isEmpty triggered and [AwaitableSlice.IsClosed] returning true
	//	- state changes to drain to closed states are single-transition irreversible
	isCloseInvoked atomic.Bool
	// isEmpty is triggered upon the slice entering closed state by being
	// read to end after Close
	//	- closed state means the queue was or became empty after
	//		[AwaitableSlice.Close] was invoked
	//	- state changes to drain and closed states are single-transition irreversible
	//	- isEmpty only creates a channel if [AwaitableSlice.EmptyCh] or
	//		[AwaitableSlice.AwaitValue] is invoked
	isEmpty Awaitable
}

// AwaitableSlice is [IterableSource]
var _ IterableSource[int] = &AwaitableSlice[int]{}

// AwaitableSlice is [ClosableAllSource]
var _ ClosableAllSource[int] = &AwaitableSlice[int]{}

// AwaitableSlice is [Sink]
var _ Sink[int] = &AwaitableSlice[int]{}

// Send enqueues a single value. Thread-safe
func (s *AwaitableSlice[T]) Send(value T) {
	defer s.postSend()
	s.queueLock.Lock()

	// add to queue if no slices
	if len(s.qSos) == 0 {
		if s.queue != nil {
			s.queue = append(s.queue, value)
			// create s.queue
		} else if s.cachedInput != nil {
			// use cachedInput allocated under outputLock
			s.queue = append(s.cachedInput, value)
			s.cachedInput = nil
		} else {
			s.queue = s.makeValueSlice(value)
		}
		return // value in queue return
	}

	// add to slices
	//	- if last slice not locally created, append to new slice
	if s.isLocalSlice {
		// append to ending local slice
		var index = len(s.qSos) - 1
		s.qSos[index] = append(s.qSos[index], value)
		return
	}

	// append local slice
	var q []T
	if s.cachedInput != nil {
		q = append(s.cachedInput, value)
		s.cachedInput = nil
	} else {
		q = s.makeValueSlice(value)
	}
	pslices.SliceAwayAppend1(&s.qSos, &s.qSos0, q)
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
	s.qSos = append(s.qSos, values)
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

// DataWaitCh returns a channel that is open on empty and closes once values are available
//   - each DataWaitCh invocation may return a different channel value
//   - Thread-safe
func (s *AwaitableSlice[T]) DataWaitCh() (ch AwaitableCh) {

	// this may initialize the cyclic awaitable
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

// EmptyCh returns an awaitable closing channel that closes once
// [AwaitableSlice.Close] invoked and the queue being or
// becoming empty
//   - EmptyCh always returns the same channel value
//   - thread-safe
func (s *AwaitableSlice[T]) EmptyCh() (ch AwaitableCh) { return s.isEmpty.Ch() }

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
	if !s.hasData() {
		s.isEmpty.Close()
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

	// fast check outside lock
	if !s.hasData() {
		return
	}

	var checkedQueue bool
	defer s.postGet(&hasValue, &checkedQueue)
	s.outputLock.Lock()

	// check inside lock
	if !s.hasData() {
		return
	}
	// there is at least one value in queueLock or outputLock

	// if s.output is empty, try transfering s.outputs[0] to s.output
	if len(s.output) == 0 && len(s.outputs) > 0 {
		// possibly save output to cachedOutput
		s.ensureSize()
		if c := cap(s.output); c <= s.maxRetainSize.Load() && s.cachedOutput == nil {
			s.cachedOutput = s.output0
		}
		// write new s.output
		var so = s.outputs[0]
		s.output = so
		s.output0 = so
		s.outputs[0] = nil
		s.outputs = s.outputs[1:]
	}
	// if output has any items, s.output is non-empty

	// try output
	if hasValue = len(s.output) > 0; hasValue {
		value = s.output[0]
		var zeroValue T
		s.output[0] = zeroValue
		s.output = s.output[1:]
		return // got value from s.output
	}
	// outputLock is empty

	// transfer from queueLock
	var slice, _ = s.sliceFromQueue(getValue)
	checkedQueue = true
	if hasValue = len(slice) > 0; !hasValue {
		return // no value available return
	}
	// non-empty slice from queueLock

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

	// fast check outside lock
	if !s.hasData() {
		return
	}

	var hasValue, checkedQueue bool
	defer s.postGet(&hasValue, &checkedQueue)
	s.outputLock.Lock()

	// check inside lock
	if !s.hasData() {
		return
	}
	// there is at least one value in queueLock or outputLock

	// try output
	if len(s.output) > 0 {
		values = s.output
		hasValue = true
		s.output0 = nil
		s.output = nil
		return
	}

	// try s.outputs
	var so = s.outputs
	if len(so) > 0 {
		values = so[0]
		hasValue = true
		so[0] = nil
		s.outputs = so[1:]
		return
	}

	// transfer from queueLock
	//	- values may be nil
	values, _ = s.sliceFromQueue(getSlice)
	checkedQueue = true

	return
}

// GetAll returns a single slice of all unread values in the queue
//   - values nil: the queue is empty
//   - thread-safe
func (s *AwaitableSlice[T]) GetAll() (values []T) {

	// fast check outside lock
	if !s.hasData() {
		return
	}

	// GetAll always sets hasData to false
	//	- disable postGet action
	var checkedQueue = true
	defer s.postGet(&checkedQueue, &checkedQueue)
	s.outputLock.Lock()

	// check inside lock
	if !s.hasData() {
		return
	}
	// there is at least one value in queueLock or outputLock

	// move any data from queueLock to outputLock
	//	- this way there is no slice alloc while holding queueLock
	s.sliceFromQueue(getNothing)
	// sliceFromQueue reset queueLockHasDataBit
	//	- if it is still zero, also reset hasDataBit using CompareAndSwap
	s.setOutputLockEmpty()
	// there is data in outputLock

	// single-slice case
	if len(s.outputs) == 0 {
		values = s.output
		s.output0 = nil
		s.output = nil
		return // s.output return
	}

	// aggregate outputLock data
	// output is a copy of s.output since preAlloc may destroy it
	var getAllSize = len(s.output)
	for _, o := range s.outputs {
		getAllSize += len(o)
	}
	// size is now length of the returned slice

	// make GetAll returned slice
	values = make([]T, getAllSize)
	var v = values
	var n int

	// s.output is not empty
	n = copy(v, s.output)
	pslices.SetLength(&s.output, 0)
	v = v[n:]

	// s.outputs
	if len(s.outputs) > 0 {
		for i := range len(s.outputs) {
			n = copy(v, s.outputs[i])
			v = v[n:]
		}
		pslices.SetLength(&s.outputs, 0)
	}

	return
}

var _ io.Reader

// Read makes AwaitableSlice [io.Reader]
//   - p: buffer to read items to
//   - n: number of items read
//   - err: possible [io.EOF] once closed and read to empty
//   - — EOF may be returned along with read items
//   - —
//   - useful if a limited number of items is to be read
//   - may be less efficient by discarding internal slices
//   - combines data transfer with close which the slice otherwise treats separately
//   - if further data is written to slice after EOF, Read provides additional data after returning EOF
func (s *AwaitableSlice[T]) Read(p []T) (n int, err error) {

	// fast check outside lock
	if !s.hasData() {
		if s.isCloseInvoked.Load() {
			err = io.EOF
		}
		return // slice empty return
	} else if len(p) == 0 {
		return // buffer zero-length return
	}
	// slice was observed to have data and buffer is of non-zero length

	// Read will update hasData so disable postGet action
	var checkedQueue = true
	defer s.postGet(&checkedQueue, &checkedQueue)
	s.outputLock.Lock()

	// check inside lock
	if !s.hasData() {
		if s.isCloseInvoked.Load() {
			err = io.EOF
		}
		return // slice empty return
	}
	// there is data in queueLock or outputLock: data will be read

	// whether p has been completely read
	var isDone bool

	// iterate i: 0, 1
	for i := range 2 {

		// read outputLock data
		if !isDone {
			isDone = s.copyToSlice(&s.output, &p, &n)
			if !isDone {
				for len(s.outputs) > 0 {
					isDone = s.copyToSlice(&s.outputs[0], &p, &n)
					if len(s.outputs[0]) == 0 {
						s.outputs = s.outputs[1:]
					}
					if isDone {
						break
					}
				}
			}
		}
		if i != 0 {
			continue
		}

		// case before accessing queueLock
		var outputIsEmpty = len(s.output) == 0 && len(s.outputs) == 0
		if isDone && !outputIsEmpty {
			// read complete and hasData does not need update
			return // Read complete return
		}
		// transfer data from queueLock and update hasData
		s.sliceFromQueue(getMax, len(p))
	}
	if !s.hasData() && s.isCloseInvoked.Load() {
		err = io.EOF
	}
	return
}

// AwaitValue awaits value or close, blocking until either event
//   - hasValue true: value is valid, possibly the zero-value like
//     a nil interface value
//   - hasValue: false: the stream is closed
//   - stream: an awaitable possibly closable source type like [Source1]
//   - — stream’s DataWaitCh Get and if present EmptyCh methods are used
//   - — stream cannot be eg. [AtomicError] because it is not awaitable
//   - AwaitValue wraps a 10-line read operation as a two-value expression
func (s *AwaitableSlice[T]) AwaitValue() (value T, hasValue bool) {

	// loop until value or closed
	var endCh AwaitableCh
	var dataWait *CyclicAwaitable
	for {

		// competing with other threads for values
		//	- may receive nothing
		if value, hasValue = s.Get(); hasValue {
			return // value read return: hasValue true, value valid
		}

		if endCh == nil {
			endCh, dataWait = s.getAwait()
		}
		select {
		case <-endCh:
			return // closable is closed return: hasValue false
		case <-dataWait.Ch():
		}
	}
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

// Seq allows for AwaitableSlice to be used in a for range clause
//   - each value is provided to yield
//   - iterates until yield retuns false or
//   - the slice was empty and in drain-close states
//   - thread-safe
//
// Usage:
//
//	for value := range awaitableSlice.Seq {
//	  value…
//	}
//	// the AwaitableSlice was empty and in drain-closed state
func (s *AwaitableSlice[T]) Seq(yield func(value T) (keepGoing bool)) {

	// loop until value or closed
	var endCh AwaitableCh
	var dataWait *CyclicAwaitable
	var value T
	var hasValue bool
	for {

		// competing with other threads for values
		//	- may receive nothing
		if value, hasValue = s.Get(); hasValue {
			if !yield(value) {
				return // consumer canceled iteration return
			}
			continue
		}

		if endCh == nil {
			endCh, dataWait = s.getAwait()
		}
		select {
		case <-endCh:
			return // closable is closed return: hasValue false
		case <-dataWait.Ch():
		}
	}
}

// SetSize set initial allocation size of slices
//   - size: allocation size in T elements for new value-slices
//   - size < 1: default allocation size 10
//   - maxRetainSize is set 100 or size if larger than 100
//   - thread-safe
func (s *AwaitableSlice[T]) SetSize(size int) {

	if size < 1 {
		size = defaultSize /* 10 */
	}
	s.size.Store(size)

	var maxSize int
	if size > maxForPrealloc /* 100 */ {
		maxSize = size
	} else {
		maxSize = maxForPrealloc
	}
	s.maxRetainSize.Store(maxSize)
}

// makeValueSlice returns an allocated slice capacity SetSize default 10
// empty slice returns a new slice of length 0 and configured capacity
//   - value: if present, is added to the new slice
//   - newSlice an allocated slice, length 0 or 1 if value present
func (s *AwaitableSlice[T]) makeValueSlice(value ...T) (newSlice []T) {

	// ensure size sizeMax are initialized
	var size = s.size.Load()
	if size == 0 {
		s.SetSize(0)
		size = s.size.Load()
	}

	// create slice optionally with value
	newSlice = make([]T, len(value), size)
	if len(value) > 0 {
		newSlice[0] = value[0]
	}

	return
}

// sliceFromQueue fetches items from queue and moves all data to output.
// Upon return, all data is with outputLock and hasData is up-to-date
//   - action getValue: seeking single value. If slice is longer, it should be saved to s.output
//   - action getSlice: seeking entire slice
//   - action getNothing: returned slice is nil. Action simply updates up-to-date
//   - action getMax: seeking maxCount number of items
//     If slice is longer, it should be saved to s.output.
//     Items should be fetched both from slice and s.outputs
//   - maxCount: optional number of items for action getMax
//   - slice: any non-empty slice from queueLock.
//     slice is not stored in outputLock.
//     Any additional data in queue is moved to s.outputs
//   - —
//   - on invocation:
//   - — output is empty
//   - — outputLock is held
//   - — hasData is true
//   - to reduce queueLock contention, sliceFromQueue:
//   - — transfers all values to output and
//   - — transfers pre-allocated slices to queue
func (s *AwaitableSlice[T]) sliceFromQueue(action getAction, maxCount ...int) (slice []T, slices [][]T) {

	// only access queueLock if there is data behind it
	if s.queueLockIsEmpty() {
		// no data behind queueLock
		return // entire queue empty return
	}

	// prealloc outside queueLock
	s.preAlloc()
	defer s.queueLock.Lock().Unlock()

	// because both locks are held, hasDataBits cam be written directly
	//	- clear queueLockHasDataBit since queueLock is emptied now
	//	- keep hasDataBit set for now
	s.hasDataBits.Store(hasDataBit) // sliceFromQueue
	s.transferCached()

	// three tasks while holding queueLock:
	//	- find what slices to return
	//	- transfer all other slices to outputLock
	//	- update hasData

	// output and outputLock is empty

	// try to get data to slice
	switch action {
	case getValue, getSlice:

		// retrieve queue if non-empty
		if len(s.queue) > 0 {
			slice = s.queue
			s.queue = nil
		}

		// if slice empty, try first of slices
		if len(slice) == 0 && len(s.qSos) > 0 {
			slice = s.qSos[0]
			s.qSos[0] = nil
			s.qSos = s.qSos[1:]
		}
	case getAll:

		// retrieve queue if non-empty
		if len(s.queue) > 0 {
			slice = s.queue
			s.queue = nil
		}

		// retrieve slices if non-empty
		if len(s.qSos) > 0 {
			slices = s.qSos
			s.qSos = nil
			s.qSos0 = nil
		}
	case getNothing, getMax:

		// getNothing: transfer any non-empty queue
		if len(s.queue) > 0 {
			var s0 = s.queue
			// transfer output0 to queueLock
			s.queue = s.output0
			s.output = s0
			s.output0 = s0
		}
	default:
		panic(perrors.ErrorfPF("Bad action: %d", action))
	}

	// possibly transfer pre-made output0 to queueLock
	if s.queue == nil {
		// transfer output0 to queueLock
		s.queue = s.output0
		s.output0 = nil
	}

	// transfer any remaining slices
	if len(s.qSos) > 0 {
		pslices.SliceAwayAppend(&s.outputs, &s.outputs0, s.qSos)
		// empty and zero-out s.slices
		pslices.SetLength(&s.qSos, 0)
		s.qSos = s.qSos0[:0]
	}
	// queueLock is now empty

	// hasData must be updated while holding queueLock
	//	- hasData is currently true
	if len(s.outputs) > 0 {
		return // slices in outputs mean data still available: no change
	}

	switch action {
	case getValue:
		// if fetching single value and more than one value in that slice,
		// not end of data: hasData should remain true
		if len(slice) > 1 {
			return // fetching multiple values return: hasData true
		}
	case getNothing:
		// if output is not empty, hasData should remain true
		if len(s.output) > 0 {
			return // values in outputLock return: hasData true
		}
	case getSlice:
		// slice will be consumed entirely
		//	- there are no more slices, so hasData is now false
	case getMax:

		// looking for specific number of items: n
		var n int
		if len(maxCount) > 0 {
			n = maxCount[0]
		}

		// if n goes negative, hasData should be true
		//	- return keeps hasData true
		//	- if s.output, s.outputs contains more than n: hasData true
		//	- if s.output, s.outputs contains less or equal to n: hasData false
		if n -= len(s.output); n < 0 {
			return // more items than n: hasData true
		}
		for i := range len(s.outputs) {
			if n -= len(s.outputs[i]); n < 0 {
				return // more items than n: hasData true
			}
		}
	default:
		panic(perrors.ErrorfPF("Bad action: %d", action))
	}
	// queueLock and outputLock are both empty

	// there is not enough data for the queue to not become empty
	// update hasDataBits while holding queueLock
	s.hasDataBits.Store(emptyNoBits) // sliceFromQueue

	return // hasData now false return
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
		if !s.hasData() {
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
	if s.hasData() == s.dataWait.Cyclic.IsClosed() {
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
	var hasData = s.hasData()

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
//const onlyCachedTrue = true

// preAlloc ensures that output0 and cachedOutput are allocated
// to configured value-slice size
//   - must hold outputLock
//   - onlyCached missing or false: both output and cachedOutput are
//     ensured to be allocate and reusable
//   - onlyCached true: by former GetAll: only cachedOutput is ensured
//     to be allocated ansd reusable
//   - purpose is to reduce allocations while holding queueLock
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
			var so = s.makeValueSlice()
			s.output0 = so
			s.output = so
		}
	}

	// cachedOutput second pre-allocation
	//	- possibly have ready for transfer
	//	- configured size may be large so only for defaultSize
	if s.cachedOutput == nil && size == defaultSize {
		s.cachedOutput = s.makeValueSlice()
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
//   - gotValue: true if the Get GetSlice GetAll Read operation retrieved a value
//   - checkedQueue: true if queueLock data was checked
//   - —
//   - The reason postGet has goValue is that if outputLock was emptied,
//     queueLock must be checked to update hasData
//   - expensive action is taken on gotValue true, checkedQueue false
//     and output empty
//   - postGet aggregates deferred actions to reduce latency
//   - postGet is invoked while holding outputLock
func (s *AwaitableSlice[T]) postGet(gotValue, checkedQueue *bool) {

	// if a value was retrieved, hasData may need update
	//	- if gotValue is true, hasData was true
	//	- — if both outputLock and queueLock are now empty
	//	- — then hasData must be set to false
	//	- if queueLock was checked, hasData was already updated
	//	- used by Get GetSlice
	if *gotValue && !*checkedQueue {
		// hasData is true
		// outputLock may be empty

		// check if there are any more values in outputLock
		if len(s.output) == 0 && len(s.outputs) == 0 {
			// if queueLockHasData is false,
			// hasData should be false
			s.setOutputLockEmpty() //postGet
		}
	}

	// reliquish outputLock and initiate eventually consistent data/close update
	s.outputLock.Unlock()
	s.updateWait() // postGet
}

// postSend sets hasData to true, relinquishes queueLock and
// initializes eventual update of DataWaitCh and EmptyCh
//   - aggregates deferred actions to reduce latency
//   - invoked while holding queueLock
func (s *AwaitableSlice[T]) postSend() {
	// set both data bits atomically
	s.hasDataBits.Store(hasDataBit | queueLockHasDataBit) // postSend
	s.queueLock.Unlock()
	s.updateWait() // postSend
}

// getAwait returns
//   - endCh: closing on Close-empty
//   - dataWait: cyclic with data wait channel
func (s *AwaitableSlice[T]) getAwait() (endCh AwaitableCh, dataWait *CyclicAwaitable) {
	endCh = s.EmptyCh()

	// ensure data-wait is initialized
	if !s.dataWait.IsActive.Load() {
		// winner thread does updateWait
		s.DataWaitCh()
	}
	dataWait = &s.dataWait.Cyclic // AwaitValue

	return
}

// copyToSlice does slice-away from src, copying to dst and adding to *np
//   - src: pointer to source-slice
//   - dst: pointer to Read p-slice buffer
//   - np: pointer to Read n integer
//   - isDone: true if dst was filled
func (s *AwaitableSlice[T]) copyToSlice(src, dst *[]T, np *int) (isDone bool) {

	// copy if anything to copy
	var d = *dst
	var sc = *src
	var nCopy = copy(d, sc)
	if nCopy == 0 {
		return // nothing to copy: isDone false
	}
	// items were copied

	// update n *src *dst isDone
	*np += nCopy
	*src = sc[nCopy:]
	*dst = d[nCopy:]
	isDone = len(d) == nCopy

	return // bytes were copied return
}

// hasData returns true if there is data in the AwaitableSlice
//   - invoked at any time, no lock required
func (s *AwaitableSlice[T]) hasData() (dataYes bool) {
	return s.hasDataBits.Load()&hasDataBit != 0
}

// queueLockIsEmpty returns true if the AwaitableSlice is now empty
//   - invoked while holding outputLock
//   - isEmpty true: the entire queue is empty,
//     hasDataBit was reset using CompareAndSwap
//   - isEmpty false: there is data in queueLock needing to be retrieved.
//     hasDataBits was not changed
//   - —
//   - invoked behind outputLock when outputLock became empty
//   - hasDataBit is known to be set
//   - purpose is to determine whether queueLock must be accessed
//   - if queueLock is also empty, the entire queue is empty
func (s *AwaitableSlice[T]) queueLockIsEmpty() (isEmpty bool) {
	for {

		// read initial state to be able to do CompareAndSwap
		var oldBits = s.hasDataBits.Load()
		// if queueLock has no data, the entire queue is empty: isEmpty true
		isEmpty = oldBits&queueLockHasDataBit == 0
		// if the queue is not empty, return to access data behind qqueueLock
		if !isEmpty {
			return // data in queueLock return
		}
		// if queueLockHasDataBit is still zero, set hasDataBit to zero, too
		var newBits hasDataBitField = emptyNoBits // entire queue is empty
		if s.hasDataBits.CompareAndSwap(oldBits, newBits) {
			return // isEmpty true return
		}
	}
}

// setOutputLockEmpty clears hasDataBit if queueLockHasDataBit is
// still cleared
//   - invoked while holding queueLock
//   - invoked while hasDataBit is set
func (s *AwaitableSlice[T]) setOutputLockEmpty() (isEmpty bool) {
	for {

		// read initial state to be able to do CompareAndSwap
		var oldBits = s.hasDataBits.Load()
		if oldBits&queueLockHasDataBit != 0 {
			// queueLock received data, the queue will not become empty
			return // no hasDataBit clear return
		} else if s.hasDataBits.CompareAndSwap(oldBits, emptyNoBits) {
			return // set to empty return
		}

	}
}

const (
	// default allocation size for new slices if Size is < 1
	defaultSize = 10
	// scavenging: max size for slice preallocation
	maxForPrealloc = 100
)

const (
	// sliceFromQueue is invoked to get a single value
	getValue getAction = iota
	// sliceFromQueue is invoked to get a non-empty slice of values
	getSlice
	// sliceFromQueue is invoked to update hasData
	getNothing
	// want maxCount elements
	getMax
	// get all non-empty slices
	getAll
)

// action for sliceFromQueue
//   - getValue getSlice getNothing getMax
type getAction uint8

const (
	// hasData indicates that the AwaitableSlice is not empty
	hasDataBit hasDataBitField = 1 << iota
	// queueLockHasDataBit indicates that there is data behind queueLock
	queueLockHasDataBit
	// emptyNoBits resets to initial zero-value state all bits cleared
	emptyNoBits hasDataBitField = 0
)

// [hasDataBit] [queueLockHasDataBit]
type hasDataBitField uint32
