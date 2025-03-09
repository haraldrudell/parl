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
	// queue is a locally made slice for individual values
	//	- behind queueLock
	//	- not a slice-away slice
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
	//	- atomic allows for updateDataWait to be invoked
	//		without any locks held
	//	- written behind queueLock
	hasData atomic.Bool
	// a pre-allocated slice for queue
	//	- behind queueLock
	//	- allocated by Get Get1 GetAll prior to acquiring queueLock
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
	// a pre-allocated slice for queue
	//	- behind outputLock
	//	- allocated by Get Get1 GetAll prior to acquiring queueLock
	cachedOutput []T
	// dataWait provides a channel that is closed upon slice empty
	//	- channel returned by [AwaitableSlice.DataWaitCh]
	//	- upon each transition empty to non-empty, channel is re-initialized
	//	- lazy initialization until any source-sink operation or
	//		DataWaitCh AwaitValue
	dataWait LazyCyclic
	// isEmptyWait is triggered upon starting the drain-close phases
	//	- ie. [AwaitableSlice.EmptyCh] invoked without argument
	//	- state changes to drain and closed states are irreversible
	//	- [AwaitableSlice.IsClosed] returns true after drain-close
	//		started and the slice being or becoming empty
	//	- between isEmptyWait triggered and isEmpty triggered,
	//		close state is tracked
	isEmptyWait Awaitable
	// isEmpty is triggered upon the slice in close state
	//	- close means [AwaitableSlice.EmptyCh] was invoked without argument
	//		and the slice was or became empty
	//	- close means [AwaitableSlice.IsClosed] returns true
	//	- state changes to drain and closed states are irreversible
	//	- initialization of isEmpty is deferred until isEmptyWait triggered
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

// DataWaitCh returns a channel that is open on empty and closes once values are available
//   - each DataWaitCh invocation may return a different channel value
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
		return // another thread initialized dataWait
	}

	// set initial state of dataWait
	s.updateWait()

	return
}

// EmptyCh returns an awaitable channel that closes on queue being or
// becoming empty
//   - doNotInitialize missing: enable closing of ch which will happen as soon
//     as the slice is empty, possibly prior to return
//   - doNotInitialize CloseAwaiter: obtain the channel but do not enable it closing.
//     A subsequent invocation with doNotInitialize missing will enable its closing thus
//     act as a deferred Close function
//   - EmptyCh always returns the same channel value
//   - thread-safe
func (s *AwaitableSlice[T]) EmptyCh(doNotClose ...CloseStrategy) (ch AwaitableCh) {

	// ch is a channel is closed or closes upon queue becoming empty
	//	- this invocation may initialize isEmpty awaitable
	ch = s.isEmpty.Ch()

	// if previously invoked, no need for initialization
	if (len(doNotClose) > 0 && doNotClose[0] == CloseAwaiter) ||
		s.isEmptyWait.IsClosed() {
		return // already closed or do not change close-state return
	}

	// establish proper state
	//	- data wait ch now in use
	if !s.isEmptyWait.Close() {
		return
	}

	// set initial state
	if !s.hasData.Load() {
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
	if !s.hasData.Load() {
		return
	}
	var checkedQueue bool
	defer s.postGet(&hasValue, &checkedQueue)
	s.outputLock.Lock()

	// check inside lock
	if !s.hasData.Load() {
		return
	}
	// there is at least one value in queueLock or outputLock

	// if output is empty, try transfering outputs[0] to output
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
	var slice = s.sliceFromQueue(getValue)
	checkedQueue = true
	if hasValue = len(slice) > 0; !hasValue {
		return // no value available return
	}
	// got slice from queueLock

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
	if !s.hasData.Load() {
		return
	}
	var hasValue, checkedQueue bool
	defer s.postGet(&hasValue, &checkedQueue)
	s.outputLock.Lock()

	// check inside lock
	if !s.hasData.Load() {
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
	values = s.sliceFromQueue(getSlice)
	checkedQueue = true

	return
}

// GetAll returns a single slice of all unread values in the queue
//   - values nil: the queue is empty
//   - thread-safe
func (s *AwaitableSlice[T]) GetAll() (values []T) {
	// fast check outside lock
	if !s.hasData.Load() {
		return
	}
	var checkedQueue = true
	defer s.postGet(&checkedQueue, &checkedQueue)
	s.outputLock.Lock()

	// check inside lock
	if !s.hasData.Load() {
		return
	}
	// there is at least one value in queueLock or outputLock

	// aggregate outputLock data
	// output is a copy of s.output since preAlloc may destroy it
	var size = len(s.output)
	for _, o := range s.outputs {
		size += len(o)
	}

	// aggregate queueLock data
	s.preAlloc(onlyCachedTrue)
	defer s.queueLock.Lock().Unlock()
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
	if !s.hasData.Load() {
		if s.isEmptyWait.IsClosed() {
			err = io.EOF
		}
		return // slice empty return
	} else if len(p) == 0 {
		return // buffer zero-length return
	}
	// slice was observed to have data and buffer is of non-zero length

	var checkedQueue = true
	defer s.postGet(&checkedQueue, &checkedQueue)
	s.outputLock.Lock()

	// check inside lock
	if !s.hasData.Load() {
		if s.isEmptyWait.IsClosed() {
			err = io.EOF
		}
		return // slice empty return
	}
	// the slice has data and buffer is non-zero length: data will be read
	// data is behing queueLock or outputLock

	// whether p has been completely read
	var isDone bool

	// iterate i: 0, 1
	for i := range 2 {

		// read outputLock data
		if !isDone {
			isDone = s.copyToSlice(&s.output, &p, &n)
			if !isDone {
				for len(s.outputs) > 0 {
					var isDone = s.copyToSlice(&s.outputs[0], &p, &n)
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
	if !s.hasData.Load() && s.isEmptyWait.IsClosed() {
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

	// endCh awaits close
	//	- nil if EmptyCh not initialized
	var endCh AwaitableCh
	// emptyInitCh awaits EmptyCh invocation
	//	- nil if already initialized
	var emptyInitCh = s.isEmptyWait.Ch()
	if !s.isEmptyWait.IsClosed() {
		emptyInitCh = nil
		endCh = s.isEmpty.Ch()
	}

	// ensure data-wait is initialized
	if !s.dataWait.IsActive.Load() {
		s.DataWaitCh()
	}
	var dataWait = &s.dataWait.Cyclic

	// loop until value or closed
	for {
		select {
		case <-emptyInitCh:
			emptyInitCh = nil
			endCh = s.isEmpty.Ch()
		case <-endCh:
			return // closable is closed return: hasValue false
		case <-dataWait.Ch():
			// competing with other threads for values
			//	- may receive nothing
			if value, hasValue = s.Get(); hasValue {
				return // value read return: hasValue true, value valid
			}
		}
	}
}

// IsClosed returns true if closable is closed or triggered
//   - isClosed is a single boolean value usable with for or if
//   - IsClosed wraps a 6-line read into a single-value boolean expression
func (s *AwaitableSlice[T]) IsClosed() (isClosed bool) {

	// if emptyCh has not been initialized,
	// the slice is not closed
	if !s.isEmptyWait.IsClosed() {
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

	// endCh is deferred initialization waiting for slice empty
	var endCh AwaitableCh
	for {

		// try obtaining value
		if s.hasData.Load() {
			// the slice is not empty
			//	- get value in competition with other threads
			if v, hasValue := s.Get(); hasValue {
				if !yield(v) {
					return // consumer canceled iteration return
				}
				continue
			}
		}
		// the slice was empty

		//	- wait until the slice has data or
		//	- the slice closes
		// atomic-performance check for channel end
		if s.isEmptyWait.IsClosed() {
			// slice was empty and in drain or close state
			return // end of values iteration end
		}
		// the slice was empty and was not in drain or close states

		// await more data or transition to drain-close states

		// initialize endCh to detect state change to drain-close
		if endCh == nil {
			// get endCh without initializing close mechanic
			endCh = s.EmptyCh(CloseAwaiter)
		}
		select {

		// await data, possibly initializing dataWait
		case <-s.DataWaitCh():
			// data was detected, try to retrieve it
			// in compeition with other threads

			// await close and end of data
		case <-endCh:
			// the slice is empty and
			// the slice transiotioned to drain-close states
			return // slice empty and in drain-close state
		}
		// there was data available
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

// sliceFromQueue fetches a value, a slice of values from queue to output
// or updates hasData. Upon return all data is with outputLock and hasData is updated
//   - action getValue: seeking single value. Longer returned slice is saved as s.output
//   - action getSlice: seeking entire slice
//   - action getNothing: slice is nil
//   - slice: any non-empty slice from queueLock, slice is not stored in outputLock
//   - upon sliceFromQueue invocation:
//   - — outputLock is empty
//   - — outputLock is held
//   - — hasData is true
//   - to reduce queueLock contention, sliceFromQueue:
//   - — transfers all values to outputLock and
//   - — transfers pre-allocated slices to queueLock
func (s *AwaitableSlice[T]) sliceFromQueue(action getAction, maxCount ...int) (slice []T) {
	// prealloc outside queueLock
	s.preAlloc()
	defer s.queueLock.Lock().Unlock()
	s.transferCached()

	// three tasks while holding queueLock:
	//	- find what slice to return
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
		if len(slice) == 0 && len(s.slices) > 0 {
			slice = s.slices[0]
			s.slices[0] = nil
			s.slices = s.slices[1:]
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
	if len(s.slices) > 0 {
		pslices.SliceAwayAppend(&s.outputs, &s.outputs0, s.slices)
		// empty and zero-out s.slices
		pslices.SetLength(&s.slices, 0)
		s.slices = s.slices0[:0]
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
		// looking for specific number of items
		//	- if s.output, s.outputs is less, hasData false
		// get how many items are sought
		var n int
		if len(maxCount) > 0 {
			n = maxCount[0]
		}
		// subtract all available items
		n -= len(s.output)
		if n >= 0 {
			for i := range len(s.outputs) {
				n -= len(s.outputs[i])
				if n < 0 {
					break
				}
			}
		}
		if n < 0 {
			return // values in outputLock return: hasData true
		}
	default:
		panic(perrors.ErrorfPF("Bad action: %d", action))
	}
	// queueLock and outputLock are both empty

	// the queue is empty: update hasData while holding queueLock
	s.hasData.Store(false)

	return // hasData now false return
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
	} else if !s.isEmptyWait.IsClosed() {
		return // neither is active
		// close emptyCh if empty
	} else if s.hasData.Load() || s.isEmpty.Close() {
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
			if s.isEmptyWait.IsClosed() {
				s.isEmpty.Close()
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
	s.isEmpty.Close()
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
//   - gotValue: true if the Get GetSlice GetAll Read operation retrieved a value
//   - — on true, if queue wsn’t checked, queue has to be checked
//   - checkedQueue: true if queueLock data was checked
//   - —
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
	if *gotValue && !*checkedQueue {

		// check if there are any more values in outputLock
		if len(s.output) == 0 && len(s.outputs) == 0 {
			// acquire queueLock to update hasData
			s.sliceFromQueue(getNothing)
		}
	}

	// reliquish outputLock and initiate eventually consistent data/close update
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
)

// action for sliceFromQueue
//   - getValue getSlice getNothing getMax
type getAction uint8
