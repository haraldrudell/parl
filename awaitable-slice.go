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
)

// AwaitableSlice is a queue as thread-safe awaitable unbound slice or slice of slices
//   - [AwaitableSlice.Send] [AwaitableSlice.Get1] allows efficient
//     transfer of single values
//   - [AwaitableSlice.SendSlice] [AwaitableSlice.Get] allows efficient
//     transfer of slices where
//     a sender relinquish slice ownership by invoking SendSlice and
//     a receiving thread gains slice ownership by invoking Get
//   - [AwaitableSlice.DataWaitCh] returns a channel that closes once data is available
//   - [AwaitableSlice.SetSize] allows for setting initial slice capacity
//   - AwaitableSlice:
//   - — can be operated with few allocations
//   - — features contention-separation between Send SendSlice and Get1 Get
//   - — offers high-throughput multiple-value operations SendSlice Get
//   - — is initialization-free awaitable and thread-less
//   - see also:
//   - — [NBChan] fully-featured unbound channel
//   - — [NBRareChan] low-usage unbound channel
type AwaitableSlice[T any] struct {
	// allocation size for new slices, effective if > 0
	//	- default 10
	size Atomic64[int]
	// queueLock makes queue and slices thread-safe
	//	- queueLock also makes Send SendSlice critical sections
	queueLock sync.Mutex
	// queue is a locally made slice for individual values
	queue []T
	// slices contains slices of values transferred by SendSlice and
	// possible subsequent locally made slices of values
	//	- slice-away slice
	slices, slices0 [][]T
	// isLocalSlice is true if the last slice of slices is locally made
	//	- only valid when slices non-empty
	isLocalSlice bool
	// written behind queueLock
	//	- indicates at all times whether the queue is empty
	hasData atomic.Bool
	// outputLock makes output thread-safe
	//	- outputLock also makes Get1 Get critical sections
	outputLock sync.Mutex
	// output is a slice being sliced away from
	output, output0 []T
	// whether DataWaitCh was invoked
	isDataWait atomic.Bool
	// allows to wait for data
	dataWaitCh CyclicAwaitable
}

// Send enqueues a single value
func (s *AwaitableSlice[T]) Send(value T) {
	s.queueLock.Lock()
	defer s.updateDataWait()
	defer s.queueLock.Unlock()
	defer s.hasData.CompareAndSwap(false, true)

	// add to queue if no slices
	if len(s.slices) == 0 {
		if s.queue == nil {
			s.queue = s.make(value)
		} else {
			s.queue = append(s.queue, value)
		}
		return
	}

	// add to slices
	//	- if last slice not locally created, append to new slice
	if !s.isLocalSlice {
		// try to save allocations on adding a slice to slicedAway s.slices
		pslices.SliceAwayAppend1(&s.slices, &s.slices0, s.make(value))
		s.isLocalSlice = true
	} else {
		// otherwise append to the local slice
		var index = len(s.slices) - 1
		s.slices[index] = append(s.slices[index], value)
	}
}

// SendSlice provides values by transferring ownership of a slice to the queue
//   - SendSlice may reduce allocations and increase performance by handling multiple values
func (s *AwaitableSlice[T]) SendSlice(values []T) {
	// ignore empty slice
	if len(values) == 0 {
		return
	}
	s.queueLock.Lock()
	defer s.updateDataWait()
	defer s.queueLock.Unlock()
	defer s.hasData.CompareAndSwap(false, true)

	// append to slices
	s.slices = append(s.slices, values)
	s.isLocalSlice = false
}

// DataWaitCh returns a channel that closes once values becomes available
func (s *AwaitableSlice[T]) DataWaitCh() (ch AwaitableCh) {
	ch = s.dataWaitCh.Ch()

	// if previously invoked, no need for initialization
	if s.isDataWait.Load() {
		return // not first invocation
	}

	// establish proper state
	// data wait ch now in use
	s.isDataWait.CompareAndSwap(false, true)
	s.updateDataWait()
	return
}

// Get1 returns one value if the queue is not empty
//   - value is valid if hasValue is true
//   - if hasValue is false, the queue is empty
//   - Get1 reduces allocations during retrieval
func (s *AwaitableSlice[T]) Get1() (value T, hasValue bool) {
	if !s.hasData.Load() {
		return
	}
	s.outputLock.Lock()
	defer s.updateDataWait()
	defer s.outputLock.Unlock()

	// try output
	if hasValue = len(s.output) > 0; hasValue {
		value = s.output[0]
		var zeroValue T
		s.output[0] = zeroValue
		s.output = s.output[1:]
		return
	}

	// get slice from queue
	var slice = s.sliceFromQueue()
	if hasValue = len(slice) > 0; !hasValue {
		return // no value available
	}

	// store slice as output
	s.output0 = slice
	value = slice[0]
	var zeroValue T
	slice[0] = zeroValue
	s.output = slice[1:]
	return
}

// Get returns a slice of values from the queue
//   - returns a non-empty slice at a time, not necessarily all data
//   - values nil: the queue is empty
//   - Get may increase performance by slice-at-a-time operation
//   - SendMany-Get operation is low allocation operation but requires
//     sender to allocate slices
//   - Send-Get operation causes Get to allocate the slices returned
func (s *AwaitableSlice[T]) Get() (values []T) {
	if !s.hasData.Load() {
		return
	}
	s.outputLock.Lock()
	defer s.updateDataWait()
	defer s.outputLock.Unlock()

	// try output
	if len(s.output) > 0 {
		values = s.output
		s.output0 = nil
		s.output = nil
		return
	}

	// get slice from queue
	values = s.sliceFromQueue()
	return
}

// SetSize set initial allocation size of slices
func (s *AwaitableSlice[T]) SetSize(size int) { s.size.Store(size) }

// make returns a new slice
//   - value, if present, is added to the new slice
func (s *AwaitableSlice[T]) make(value ...T) (newSlice []T) {
	newSlice = make([]T, len(value), s.sizeToUse())
	if len(value) > 0 {
		newSlice[0] = value[0]
	}
	return
}

// getSliceFromQueue fetches a slice to output from queue
//   - invoked when output empty
func (s *AwaitableSlice[T]) sliceFromQueue() (slice []T) {

	// ensure output0 is a slice of capacity sizeToUse
	//	- avoids allocation while holding queueLock
	if s.output0 == nil || cap(s.output0) != s.sizeToUse() {
		s.output0 = s.make()
	}

	// find a slice in queue
	s.queueLock.Lock()
	defer s.queueLock.Unlock()

	// take queue
	if len(s.queue) > 0 {
		slice = s.queue
		s.queue = s.output0
		s.output0 = nil
		// otherwise take slice from slices
	} else if len(s.slices) > 0 {
		slice = s.slices[0]
		s.slices[0] = nil
		s.slices = s.slices[1:]
	}

	// update hasData while holding queueLock
	if len(slice) == 0 {
		s.hasData.CompareAndSwap(true, false)
	}
	return
}

// sizeToUse returns slice size for make
func (s *AwaitableSlice[T]) sizeToUse() (size int) {
	if size = s.size.Load(); size < 1 {
		size = defaultSize
	}
	return
}

// setData updates dataWaitCh if DataWaitCh was invoked
func (s *AwaitableSlice[T]) updateDataWait() {

	// is data wait used yet?
	if !s.isDataWait.Load() {
		return // no data wait
	}

	// is data wait in the correct state already?
	// if queue has data, dataWaitCh should be closed
	if s.hasData.Load() {
		s.dataWaitCh.Close()
		return
	}

	// if queue is empty, data wait ch should be open
	s.dataWaitCh.Open()
}
