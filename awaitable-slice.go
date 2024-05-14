/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
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
//   - [AwaitableSlice.Send] [AwaitableSlice.Get1] allows efficient
//     transfer of single values
//   - [AwaitableSlice.SendSlice] [AwaitableSlice.Get] allows efficient
//     transfer of slices where
//     a sender relinquish slice ownership by invoking SendSlice and
//     a receiving thread gains slice ownership by invoking Get
//   - [AwaitableSlice.DataWaitCh] returns a channel that closes once data is available
//     making the queue awaitable
//   - [AwaitableSlice.SetSize] allows for setting initial slice capacity
//   - AwaitableSlice benefits:
//   - — is trouble-free data-sink: non-blocking-unbound-send non-deadlocking panic-free error-free
//   - — is initialization-free awaitable thread-less thread-safe
//   - — features channel-based wait usable with Go select and default
//   - — is unbound configurable low-allocation
//   - — features contention-separation between Send SendSlice and Get1 Get
//   - — offers high-throughput multiple-value operations SendSlice Get
//   - — avoids temporary large-slice memory leaks by using size
//   - — avoids temporary memory leaks by zero-out of unused slice elements
//   - see also:
//   - — [NBChan] fully-featured unbound channel
//   - — [NBRareChan] low-usage unbound channel
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
	hasData     atomic.Bool
	cachedInput []T
	// outputLock makes output thread-safe
	//	- outputLock also makes Get1 Get critical sections
	outputLock sync.Mutex
	// output is a slice being sliced away from
	//	- behind outputLock, slice-away slice
	output, output0   []T
	outputs, outputs0 [][]T
	cachedOutput      []T
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
		if s.queue == nil {
			// new queue
			var q []T
			if s.cachedInput != nil {
				q = s.cachedInput
				s.cachedInput = nil
				q[0] = value
				s.queue = q
			} else {
				s.queue = s.make(value)
			}
		} else {
			s.queue = append(s.queue, value)
		}
		return
	}

	// add to slices
	//	- if last slice not locally created, append to new slice
	if !s.isLocalSlice {
		var q []T
		if s.cachedInput != nil {
			q = s.cachedInput
			s.cachedInput = nil
			q[0] = value
		} else {
			q = s.make(value)
		}
		// try to save allocations on adding a slice to slicedAway s.slices
		pslices.SliceAwayAppend1(&s.slices, &s.slices0, q)
		s.isLocalSlice = true
	} else {
		// otherwise append to the local slice
		var index = len(s.slices) - 1
		s.slices[index] = append(s.slices[index], value)
	}
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

func (s *AwaitableSlice[T]) EmptyCh() (ch AwaitableCh) {
	// this may initialize the cyclic awaitable
	ch = s.emptyWait.Cyclic.Ch()

	// if previously invoked, no need for initialization
	if s.emptyWait.IsActive.Load() {
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

// Get1 returns one value if the queue is not empty
//   - hasValue true: value is valid
//   - hasValue false: the queue is empty
//   - Get1 may attain allocation-free receive or allocation-free operation
//   - Thread-safe
func (s *AwaitableSlice[T]) Get1() (value T, hasValue bool) {
	if !s.hasData.Load() {
		return
	}
	defer s.postGet()
	s.outputLock.Lock()

	// if output empty, fetch it from outputs
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
	if value, hasValue = s.fetch1(&s.output); hasValue {
		return // got value from s.output
	}

	// fetch values from queue
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

// Get returns a slice of values from the queue
//   - values non-nil: a non-empty slice at a time, not necessarily all data.
//     values is never non-nil and empty
//   - — if data arrives via Send, each Get empties the queue
//   - — if data arrives via SendMany, each Get receives one SendMany slice
//   - values nil: the queue is empty
//   - Get may increase performance by slice-at-a-time operation
//   - SendMany-Get operation is low allocation operation but requires
//     sender to allocate slices
//   - Send-Get operation causes Get to allocate the slices returned
//   - Thread-safe
func (s *AwaitableSlice[T]) Get() (values []T) {
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

	// get slice from queue
	values = s.sliceFromQueue(isSlice)

	return
}

func (s *AwaitableSlice[T]) GetAll() (values []T) {
	if !s.hasData.Load() {
		return
	}
	defer s.postGet()
	s.outputLock.Lock()

	// aggregate output data
	var size = len(s.output)
	for _, o := range s.outputs {
		size += len(o)
	}

	s.preAlloc()
	s.queueLock.Lock()
	defer s.queueLock.Unlock()
	s.transferCached()

	// aggregate queueLock data
	size += len(s.queue)
	for _, s := range s.slices {
		size += len(s)
	}

	// no data
	if size == 0 {
		return
	}
	// will return all data so queue will be empty
	s.hasData.Store(false)

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
	pslices.SliceAwayAppend(&s.outputs, &s.outputs0, nil)
	if len(s.queue) > 0 {
		values = append(values, s.queue...)
		pslices.SetLength(&s.queue, 0)
	}
	for _, s := range s.slices {
		values = append(values, s...)
	}
	pslices.SliceAwayAppend(&s.slices, &s.slices0, nil)

	return
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

func (s *AwaitableSlice[T]) fetch1(slicep *[]T) (value T, hasValue bool) {
	var slice = *slicep
	if hasValue = len(slice) > 0; hasValue {
		var p0 = &slice[0]
		value = *p0
		var zeroValue T
		*p0 = zeroValue
		*slicep = slice[1:]
	}
	return
}

// make returns a new slice
//   - value, if present, is added to the new slice
func (s *AwaitableSlice[T]) make(value ...T) (newSlice []T) {
	// ensure size sizeMax are initialized
	var size = s.size.Load()
	if size == 0 {
		s.SetSize(0)
		size = s.size.Load()
	}
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

// getSliceFromQueue fetches a slice to output from queue
//   - getSlice true: seeking entire slice
//   - getSlice false: seeking single value
//   - invoked when output empty
func (s *AwaitableSlice[T]) sliceFromQueue(getSlice bool) (slice []T) {

	// fetch slices from queue
	s.preAlloc()
	s.queueLock.Lock()
	defer s.queueLock.Unlock()
	s.transferCached()

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

	// if slice retrieved
	if len(slice) > 0 {
		// retrieve all but one from s.slices
		if len(s.slices) > 1 {
			// transfer from s.slices to s.outputs
			pslices.SliceAwayAppend(&s.outputs, &s.outputs0, s.slices[:len(s.slices)-1])
			s.slices = s.slices[len(s.slices)-1:]
			// zero out unused elements
			//	- returns slice-away s.slices to s.slices0
			pslices.SliceAwayAppend(&s.slices, &s.slices0, nil)
		}
		// if slice is non-empty this is not the last data
		if len(s.slices) > 0 {
			return
			// if fetching single value and only one value in slice,
			//	- thisis end of data
		} else if !getSlice && len(slice) > 1 {
			return // no, single-value fetch and more than one value
		}
		// fetching single value in slice length 1 or
		// the last available slice
		//	- flag end of data
		s.hasData.Store(false)
		return
	}
	// queue was empty
	//	- s.slices may have data

	// take the first slice from s.slices, if any
	if len(s.slices) > 0 {
		slice = s.slices[0]
		if len(s.slices) == 1 {
			s.slices[0] = nil
			s.slices = s.slices[1:]
		} else if len(s.slices) > 2 {
			// transfer all but 1 of remaining to s.outputs
			pslices.SliceAwayAppend(&s.outputs, &s.outputs0, s.slices[1:len(s.slices)-1])
			s.slices = s.slices[len(s.slices)-1:]
			// zero out unused elements
			pslices.SliceAwayAppend(&s.slices, &s.slices0, nil)
		}
		return
	}
	// it is end of data

	// the queue is empty: update hasData while holding queueLock
	s.hasData.Store(false)

	return
}

// setData updates dataWaitCh if DataWaitCh was invoked
func (s *AwaitableSlice[T]) updateWait() {

	// dataWait closes on data available, then re-opens
	//	- hasData true: dataWait should be closed
	//	- emptyWait closes on empty and remains closed
	//	- hasData observed false, emptyWait should be closed

	// is dataWait or emptyWait in use?
	var dataWait = s.dataWait.IsActive.Load()

	if dataWait {
		// atomic check based on dataWait
		if s.hasData.Load() == s.dataWait.Cyclic.IsClosed() {
			return // atomically state was ok
		}
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
			s.emptyWait.Cyclic.Close()
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

func (s *AwaitableSlice[T]) ensureSize() (size int) {
	// ensure size sizeMax are initialized
	if size = s.size.Load(); size == 0 {
		s.SetSize(0)
		size = s.size.Load()
	}
	return
}

func (s *AwaitableSlice[T]) preAlloc() {

	// output0 first pre-allocation
	//	- ensure output0 is a slice of good capacity
	//	- may be transferred to queueLock
	//	- avoids allocation while holding queueLock
	var size = s.ensureSize()
	// should output0 be allocated?
	var makeOutput = s.output0 == nil
	if !makeOutput {
		// check capacity of existing output
		var c = cap(s.output0)
		// only for defaultSize–maxRetainSize
		makeOutput = c < defaultSize || c > s.maxRetainSize.Load()
	}
	if makeOutput {
		var so = s.make()
		s.output0 = so
	}

	// cachedOutput second pre-allocation
	//	- possibly have ready for transfer
	//	- configured size may be large so only for defaultSize
	if s.cachedOutput == nil && size == defaultSize {
		s.cachedOutput = s.make()
	}
}

func (s *AwaitableSlice[T]) transferCached() {

	// transfer cachedOutput to queueLock
	if s.cachedInput == nil && s.cachedOutput != nil {
		s.cachedInput = s.cachedOutput
		s.cachedOutput = nil
	}
}

func (s *AwaitableSlice[T]) postGet() {
	s.outputLock.Unlock()
	s.updateWait()
}

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

func (s *AwaitableSlice[T]) postSend() {
	s.hasData.Store(true)
	s.queueLock.Unlock()
	s.updateWait()
}
