/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices/pslib"
)

// outputQueue contains dequeueing lock and structures
type outputQueue[T any] struct {
	// InQ contains lock and queues for input methods
	//	- separate input and output locks reduces contention
	InQ inputQueue[T]
	// queue provides: lock sliceList sliceList0
	// HasDataBits is bit-field indicating whether there is data behind locks
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
	HasDataBits hasData
	// list provides: lock sliceList sliceList0
	list[T]
	// head is a slice being sliced away from
	//	- head is the result of a slice-expression on head0.
	//		head and head0 share underlying array
	//	- head is slice-away slice
	head []T
	// head0 is full-length slice returned by make
	//	- the purpose of head0 is to return head to
	//		the beginning of the underlying array
	head0 []T
	// a pre-allocated value-slice transfered to outputQueue Lock
	//	- either nil or zero-length with non-zero capacity
	//	- allocated outside of outputQueue lock
	cachedOutput []T
	// lastPrimaryLarge is true if last emptyInQ encountered
	// a primary slice with length larger than allocation size
	//	- behind outQ lock
	//	- allows large slices to be transferred from outQ to InQ
	lastPrimaryLarge bool
	// sizeMax4KiB is true if Inq Size is max 4 KiB
	sizeMax4KiB atomic.Bool
}

// get returns one value
//   - hasData must have been verified true
func (o *outputQueue[T]) get() (value T) {

	// handle length
	if o.InQ.IsLength.Load() {
		o.InQ.Length.Add(-1)
	}

	// check head
	//	- at least one of outQ and InQ has data
	var wasOutQEmpty = len(o.head) == 0
	if wasOutQEmpty {

		// outQ is empty, so InQ cannot be empty
		//	- updates hasData
		var slice, _ = o.transferToOutQ(getValue)
		// for low-alloc, head could be allocated
		//	- otherwise head is unretainable or nil
		o.setHead(slice)
	}
	// head is not empty

	// fetch value using slice-away
	var outQNowEmpty bool
	value, outQNowEmpty = o.dequeueFromHead()

	// wasOutQEmpty true: hasData was already updated
	// by transferToOutQ
	//	- outQNowEmpty false: outQ is not empty
	if wasOutQEmpty || !outQNowEmpty {
		return // hasData was already updated
	}

	// outQ transitioned to empty, update hasData
	o.HasDataBits.setOutputLockEmpty()

	return
}

// getSlice returns a slice of values
//   - hasData must have been verified true
func (o *outputQueue[T]) getSlice() (values []T) {

	// try head
	if len(o.head) > 0 {

		// return the head slice
		values, _ = o.getHead()

		// try to get next head slice
		var head = o.dequeueFromList()
		// if read to empty, update hasData
		if head == nil {
			o.HasDataBits.setOutputLockEmpty()
		} else {
			// store new head slice
			o.setHead(head)
		}
	} else {

		// transfer from inQ
		//	- updates hasData
		values, _ = o.transferToOutQ(getSlice)

		// if outQ is not empty, head must be set
		var head = o.dequeueFromList()
		if head != nil {
			// for low-alloc, head could be allocated
			//	- otherwise head is unretainable or nil
			o.setHead(head)
		}
	}

	// handle length
	if o.InQ.IsLength.Load() {
		o.InQ.Length.Add(-len(values))
	}

	return
}

// getSlices empties the queue at near zero allocations
func (o *outputQueue[T]) getSlices(buffer [][]T) (slices [][]T) {

	// move any data from inQ to outQ
	o.transferToOutQ(getNothing)
	// inQ is empty, outQ is not empty

	// set slices
	var sliceListLength, _ = o.getListMetrics()
	if buffer != nil {
		slices = buffer
	} else {
		// add one for head
		slices = make([][]T, 0, sliceListLength+1)
	}

	// retrieve head slice
	//	- because outQ is not empty, head is not empty
	var head, _ = o.getHead()
	slices = append(slices, head)

	// retrieve any slices in sliceList
	if sliceListLength > 0 {
		var sliceList = o.getListSlice()
		slices = append(slices, sliceList...)
		// empty and zero-out sliceList
		o.clearList()
	}

	// track length
	if o.InQ.IsLength.Load() {
		var length int
		for i := range len(slices) {
			length += len(slices[i])
		}
		if length > 0 {
			o.InQ.Length.Add(-length)
		}
	}

	return
}

// getAll returns a single slice of all values in queue
func (o *outputQueue[T]) getAll() (values []T) {

	// move any data from inQ to outQ
	//	- this way there is no slice alloc while holding inQ lock
	o.transferToOutQ(getNothing)
	// inQ is empty, outQ is not empty

	// single-slice case
	if o.isEmptyList() {
		// slice list is empty, so output slice cannot be empty
		values, _ = o.getHead()
		return // primary output slice return
	}
	// outQ contains more than one slice

	// aggregate outQ
	// allSize is length of the returned slice >= 2
	var allSize = len(o.head) + o.getListElementCount()

	// handle length
	if o.InQ.IsLength.Load() {
		o.InQ.Length.Add(-allSize)
	}

	// make slice to return
	values = make([]T, allSize)

	// primary output slice is not empty
	var n = o.dequeueHead(values)

	// any outQ sliceList
	if !o.isEmptyList() {
		o.dequeueList(values[n:])
	}

	return
}

// read makes AwaitableSlice [io.Reader]
//   - p cannot be empty
func (o *outputQueue[T]) read(p []T) (n int, err error) {

	// read outQ head and sliceList
	//	- either outQ or InQ is non-empty
	if isDone, isOutQNowEmpty := o.dequeueNFromOutput(&p, &n); isDone {

		// if output now empty, must update hasData
		if isOutQNowEmpty {
			// possibly update hasData
			o.HasDataBits.setOutputLockEmpty()
		}
	} else {

		// transfer data from InQ
		//	- updates hasData
		o.transferToOutQ(getMax, len(p))
		// InQ empty, outQ possibly empty

		// read outQ head and sliceList
		o.dequeueNFromOutput(&p, &n)
	}

	// handle length
	if o.InQ.IsLength.Load() && n > 0 {
		o.InQ.Length.Add(-n)
	}

	return
}

// setHead sets the head slice
func (o *outputQueue[T]) setHead(head []T) {
	o.head = head
	if c := cap(head); len(head) < c {
		head = head[:c]
	}
	o.head0 = head
}

// getOutputSlice gets the output slice and its ownership
func (o *outputQueue[T]) getHead() (head, head0 []T) {
	head = o.head
	head0 = o.head0
	o.head = nil
	o.head0 = nil
	return
}

// dequeueFromHead retrieves a value from head
//   - outQNowEmpty true: outQ was read to empty
//   - —
//   - head must have been verified to not be empty
//   - slice-away dequeue
func (o *outputQueue[T]) dequeueFromHead() (value T, outQNowEmpty bool) {

	// get value with possible zero-out
	value = o.head[0]
	if o.InQ.ZeroOut.Load() != pslib.NoZeroOut {
		clear(o.head[0:1])
	}

	// if head did not become empty: done
	if len(o.head) >= 2 {
		// slice-away
		o.head = o.head[1:]
		return // value from head, head not empty return
	}
	// head was read to empty

	// see if there is new head
	var head = o.dequeueFromList()
	if head != nil {
		o.trySaveHeadToCachedOutput()
		o.setHead(head)
		return // value from head, new head not empty return
	}
	// outQ is now empty

	// reset head slice
	o.head = o.head0[:0]
	outQNowEmpty = true

	return // value from head, outQ empty return
}

// dequeueHead moves all elements from primary output
// to dest
//   - dest: destination slice
//   - — dest must be of sufficient length
//   - n: number of elements copies
func (o *outputQueue[T]) dequeueHead(dest []T) (n int) {
	n = copy(dest, o.head)
	if o.InQ.ZeroOut.Load() != pslib.NoZeroOut {
		clear(o.head)
	}
	o.head = o.head0[:0]
	return
}

// dequeueNFromOutput copies to dest from all slices
//   - dest cannot be empty
//   - useful to [io.Read]-type functions
func (o *outputQueue[T]) dequeueNFromOutput(dest *[]T, np *int) (isDone, isOutQNowEmpty bool) {

	// empty case
	if len(o.head) == 0 {
		return // outQ was empty return
	}

	// copy from head with possible zero-out
	var zeroOut = o.InQ.ZeroOut.Load()
	// read from head and sliceList
	isDone = CopyToSlice(&o.head, dest, np, zeroOut) ||
		o.dequeueNFromList(dest, np, zeroOut)

	// check new outQ state
	if len(o.head) != 0 {
		return // head not read to empty return
	}

	var head = o.dequeueFromList()
	if head == nil {
		isOutQNowEmpty = true
		o.head = o.head0[:0]
		return // OutQ read to empty return
	}

	o.trySaveHeadToCachedOutput()
	o.setHead(head)

	return
}

// trySaveHeadToCachedOutput saves the head slice as
// cachedOutput if it is suitable
//   - maxCap: maximum capacity that is retained
//   - when get fetches new head slice from sliceList,
//     the previous head slice is possibly saved
func (o *outputQueue[T]) trySaveHeadToCachedOutput() {
	if o.cachedOutput != nil || o.InQ.IsLowAlloc.Load() {
		return // cached output is already present
	} else if c := cap(o.head0); c == 0 || c > o.InQ.MaxRetainSize.Load() {
		return // unsuitable slice
	}
	// cachedOutput: zero-length non-zero capacity
	o.cachedOutput = o.head0[:0]
}

// initLength configures the queue to track length
//   - length tracking is a one-off irreversible transition
//   - thread-safe
func (o *outputQueue[T]) initLength() {
	defer o.lock.Lock().Unlock()

	// check if another thread already configures length tracking
	if o.InQ.IsLength.Load() {
		return // nothing to do return
	}
	var length = len(o.head) + o.getListElementCount()
	defer o.InQ.lock.Lock().Unlock()

	length += len(o.InQ.primary) + o.InQ.getListElementCount()
	o.InQ.Length.Store(length)
	o.InQ.MaxLength.Value(uint64(length))
	o.InQ.IsLength.Store(true)
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
func (o *outputQueue[T]) transferToOutQ(action getAction, maxCount ...int) (slice []T, slices [][]T) {

	// only access queueLock if there is data behind it
	if o.HasDataBits.isInQEmpty() {
		// no data behind queueLock
		if action != getNothing {
			return // entire queue empty return
		}

		// GetAll GetSlices completely empties outQ
		//	- if InQ still empty, hasData should be set to false
		//	- if inQ-has-data is still cleared,
		//		reset hasDataBit using CompareAndSwap
		o.HasDataBits.setOutputLockEmpty()
		return // only data in outQ return
	}
	// inQ has data, outQ is empty, hasData is true

	// slice may be head of data depending on action
	slice = o.emptyInQ(action)

	// determine and update hasData
	switch action {
	case getValue:
		// if fetching single value and more than one value in that slice,
		// not end of data: hasData should remain true
		if len(slice) > 1 || !o.isEmptyList() {
			return // have more than one value: hasData true
		}
	case getNothing:
		// getAll and getSlices will empty entire queue: hasData false

	case getSlice:
		// slice will be consumed entirely
		if !o.isEmptyList() {
			return // more slices exist: hasData true
		}
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
		if n -= len(o.head); n < 0 {
			return // more items than n: hasData true
		} else if n -= o.getListElementCount(); n < 0 {
			return // more items than n: hasData true
		}
	default:
		panic(perrors.ErrorfPF("Bad action: %d", action))
	}
	// hasData should be false

	o.HasDataBits.setOutputLockEmpty()

	return // hasData updated return
}

// emptyInQ acquires inQ lock and transfer all elements to outQ
//   - outQ lock must be held
func (o *outputQueue[T]) emptyInQ(action getAction) (slice []T) {

	// ensure outQ has sliceList when transfer from InQ
	// will require it
	if _, c := o.getListMetrics(); c == 0 {
		// outQ needs sliceList if:
		//	- InQ sliceList is not empty.
		//		inQ sliceList length is unknown outside InQ lock
		//	- action is getNothing and outQ head is not empty
		//	- outQ sliceList is only allocated once
		var needSliceList = o.InQ.HasList.Load() ||
			action == getNothing && len(o.head) > 0
		if needSliceList {
			o.setList(o.InQ.makeSliceList(minElements))
		}
	}

	// futurePrimary is a replacement-slice for InQ primary
	//   - because InQ is not empty, InQ primary will not be empty.
	//		When InQ primary is transferred,
	//		a replacement slice is required
	//	- the replacement slice may be a usable outQ head.
	//		outQ head should be verified to be reusable
	//   - as a special case,
	//		for action getNothing if outQ head is not empty,
	//     a separate slice is required as new InQ primary
	//	- futurePrimary is only nil if InQ primary should be left as nil
	var futurePrimary []T

	// futureSliceList is prealloc for inQ SliceList
	//	- inQ.HasList atomic false indicates it is required
	var futureSliceList [][]T
	if !o.InQ.IsLowAlloc.Load() {
		// always sets futurePrimary
		futurePrimary, futureSliceList = o.preAlloc()
	}
	defer o.InQ.lock.Lock().Unlock()

	// because both locks are held, hasDataBits can be written directly
	//	- clear inQHasDataBit since inQ will be emptied now
	//	- keep hasDataBit set for now
	o.HasDataBits.resetToHasDataBit()

	// transfer any pre-allocated sliceList
	if futureSliceList != nil {
		o.InQ.setList(futureSliceList)
		o.InQ.HasList.Store(true)
	}

	// transfer cachedOutput to cachedInput
	if !o.InQ.HasInput.Load() && o.cachedOutput != nil {
		o.InQ.setCachedInput(o.cachedOutput)
		o.cachedOutput = nil
	}

	// three tasks while holding InQ lock:
	//	- retrieve the returned slice value
	//	- transfer all other slices to outQ lock

	var primary = o.InQ.getPrimary()
	o.lastPrimaryLarge = len(primary) > o.InQ.Size.Load()

	// try to get data to slice
	switch action {
	case getValue, getSlice:
		slice = primary

	case getNothing, getMax:

		// transfer InQ primary to outQ head or
		// outQ sliceList
		if len(o.head) > 0 {
			o.enqueueInList(primary)

		} else {

			// save primary to outQ head
			//	- futurePrimary contains next primary
			o.setHead(primary)
		}

	default:
		panic(perrors.ErrorfPF("Bad action: %d", action))
	}

	// set new primary
	if futurePrimary != nil {
		o.InQ.setPrimary(futurePrimary)
	}

	// transfer any remaining slices
	if !o.InQ.isEmptyList() {
		var slices = o.InQ.getListSlice()
		o.enqueueInList(slices...)
		// empty and zero-out s.slices
		o.InQ.clearList()
	}

	return // inQ is now empty
}

// preAlloc ensures that head and cachedOutput are allocated
// to configured value-slice size
//   - futurePrimary: replaces InQ primary during transfer, always present
//   - futureSliceList: present when InQ does not have sliceList allocated
//   - —
//   - must hold outQ lock
//   - —
//   - onlyCached missing or false: both output and cachedOutput are
//     ensured to be allocate and reusable
//   - onlyCached true: by former GetAll: only cachedOutput is ensured
//     to be allocated and reusable
//   - purpose of preAlloc is to reduce allocations while holding InQ lock
func (o *outputQueue[T]) preAlloc() (futurePrimary []T, futureSliceList [][]T) {

	// if inQ does not have sliceList, pre-allocate
	// in futureSliceList
	if !o.InQ.HasList.Load() {
		futureSliceList = o.InQ.makeSliceList(noMinSize)
	}

	// makeOutput causes next InQ primary to be a new value-slice
	//	- presently, outQ head may be possibly allocated and non-empty
	//	- for action getNothing, outQ head slice may be non-empty.
	//		futurePrimary is then used as InQ primary
	//	- it should be used as InQ primary slice after transfer
	//	- reusing a large slice leads to temporary memory leak
	var makeOutput = len(o.head) > 0 || // action getNothing
		cap(o.head) < minElements // outQ head too small

		// check for large slice, unless last primary was large
	if !makeOutput && !o.lastPrimaryLarge {
		makeOutput = cap(o.head) > o.InQ.MaxRetainSize.Load()
	}

	if makeOutput {
		futurePrimary = o.InQ.makeValueSlice()
	} else {
		futurePrimary, _ = o.getHead()
	}

	// cachedOutput transfer to back-up cachedInput
	// InQ back-up slice
	//	- speculative allocation if value-slice size is 4 KiB or less
	if o.cachedOutput == nil && // there is no cachedOutput
		!o.InQ.HasInput.Load() && // and there is no cachedInput
		o.sizeMax4KiB.Load() { // size is max 4 KiB
		// cachedOutput: zero-length non-zero capacity
		o.cachedOutput = o.InQ.makeValueSlice()
	}

	return
}

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
