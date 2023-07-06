/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
)

// Get returns a slice of n or default all available items held by the channel.
//   - if channel is empty, 0 items are returned
//   - Get is non-blocking
//   - n > 0: max this many items
//   - n == 0 (or <0): all items
//   - Get is panic-free non-blocking error-free thread-safe
func (n *NBChan[T]) Get(elementCount ...int) (allItems []T) {

	// empty NBChan: noop return
	var unsentCount int
	if unsentCount = int(atomic.LoadUint64(&n.unsentCount)); unsentCount == 0 {
		return // no items available return: nil slice
	}

	// notify of pending Get
	n.preGet()

	// arguments
	// soughtItemCount: 0 for isAllItems, >0 for that many items
	var soughtItemCount int
	if len(elementCount) > 0 {
		if soughtItemCount = elementCount[0]; soughtItemCount < 0 {
			soughtItemCount = 0
		}
	}
	// Get request seeks all available items
	var isAllItems = soughtItemCount == 0

	if isAllItems {
		allItems = make([]T, unsentCount)[:0] // approximate size
	}

	n.outputLock.Lock()
	defer n.postGet()

	// get possible item from send thread
	if item, itemValid := n.collectSendThreadValue(); itemValid {
		allItems = append(allItems, item)
		if !isAllItems {
			if soughtItemCount--; soughtItemCount == 0 {
				return // fetch complete return
			}
		}
	}

	// fetch from n.outputQueue
	allItems = n.fetchFromOutput(&soughtItemCount, isAllItems, allItems)
	if !isAllItems && soughtItemCount == 0 {
		return // fetch complete return
	}

	// fetch from m.inputQueue
	if n.swapQueues() {
		allItems = n.fetchFromOutput(&soughtItemCount, isAllItems, allItems)
	}

	return
}

// preGet registers a pending Get invocation priot to outputLock
func (n *NBChan[T]) preGet() {
	if atomic.AddUint64(&n.gets, 1) == 1 {
		n.getsWait.HoldWaiters()
	}
}

// postGet is the deferred ending function for [NBChan.Get]
//   - decrease number of Get invocations
//   - if more Get invocations are pending, do nothing
//   - alert or launch thread if last Get
//   - release outputLock
func (n *NBChan[T]) postGet() {
	defer n.outputLock.Unlock()

	// decrement gets
	var gets = atomic.AddUint64(&n.gets, ^uint64(0))

	// update dataAvailable
	var unsentCount = atomic.LoadUint64(&n.unsentCount)
	n.setDataAvailable(unsentCount > 0)

	// is thread progress required?
	if gets > 0 || // more Get invocations pending return: do not alert or launch thread
		n.noThread.Load() { // this instance has no thread
		return // no thread progess return
	}

	// ensure thread progress

	// stop holding sendThread for pending Get invocations
	var threadIsAtGetWait = n.getsWait.Count() > 0
	n.getsWait.ReleaseWaiters()
	if threadIsAtGetWait {
		return // sendThread was notified return
	}

	// if an always-thread is idle, alert it
	//	- always-threads wait for alert when there is no data
	//	- sends and gets may keep the always thread waiting until this point
	//	- if data is now available, alert it
	if unsentCount > 0 && // data is available
		n.alertThread(nil) {
		return // always-thread progress guaranteed return
	}

	// if there is data, sendThread should be running
	//	- this thread holds outputLock, so data cannot be taken from outputBuffer
	//	- sendThread may be running and can decrease unsentCount
	//	- Send/SendMany may be ongoing adding data
	if n.isRunningThread.Load() {
		return // thread already running return
		// ensure there is data available for thread launch
	} else if len(n.outputQueue) == 0 && // output queue empty
		!n.swapQueues() { // input queue was empty
		return // no data available return
	}
	if !n.tryStartThread() {
		return // not winner return
	}

	// start the thread
	var value = n.outputQueue[0]
	n.outputQueue = n.outputQueue[1:]
	n.startThread(value)
}

// collectSendThreadValue receives any value in sendThread channel send
func (n *NBChan[T]) collectSendThreadValue() (value T, hasValue bool) {
	n.threadLock.Lock()
	defer n.threadLock.Unlock()

	if !n.isRunningThread.Load() {
		return // thread not running
	}
	select {
	case <-*n.closesOnThreadSend.Load():
	case value, hasValue = <-n.closableChan.Ch():
	}
	return
}

// discardThreadValue ends any sendThread channel send
func (n *NBChan[T]) isSendThreadChannelSend() (isChannelSendBlock bool) {
	var chp = n.closesOnThreadSend.Load()
	if chp == nil {
		return // thread not initialized return: isChannelSendBlock false
	}
	select {
	case <-*chp:
		// received from closesOnThreadSend: thread is not in send
		//	- isChannelSendBlock: false
	default:
		// closesOnThreadSend is not closed yet:
		//	- thread is headed to or in channel-send block
		isChannelSendBlock = true
	}
	return
}

// swapQueues swaps n.inputQueue and n.outputQueue0
//   - hasData true means data is available
//   - hasData false means inputQueue was empty and a swap did not take place
//   - n.outputQueue must be empty
//   - invoked while holding [NBChan.outputLock]
//   - [NBChan.inputLock] cannot be held
func (n *NBChan[T]) swapQueues() (hasData bool) {
	n.inputLock.Lock()
	defer n.inputLock.Unlock()

	if hasData = len(n.inputQueue) > 0; !hasData {
		return // no data in input queue return
	}

	// swap the queues
	n.outputQueue = n.inputQueue
	atomic.StoreUint64(&n.outputCapacity, uint64(cap(n.outputQueue)))
	n.inputQueue = n.outputQueue0
	atomic.StoreUint64(&n.inputCapacity, uint64(cap(n.inputQueue)))
	n.outputQueue0 = n.outputQueue[:0]
	return
}

// fetchFromOutput gets items from [NBChan.outputQueue]
//   - [NBChan.outputLock] must be held
func (n *NBChan[T]) fetchFromOutput(soughtItemCount *int, isAllItems bool, allItems0 []T) (allItems []T) {
	allItems = allItems0

	// empty queue case: no items
	var itemGetCount = len(n.outputQueue)
	if itemGetCount == 0 {
		return // no available items return
	}

	// entire queue case: itemCount items
	var zeroValue T
	var soughtIC = *soughtItemCount
	if isAllItems || itemGetCount <= soughtIC {
		allItems = append(allItems, n.outputQueue...)
		for i := 0; i < itemGetCount; i++ {
			n.outputQueue[i] = zeroValue
		}
		n.outputQueue = n.outputQueue[:0]
		atomic.AddUint64(&n.unsentCount, uint64(-itemGetCount))
		if !isAllItems {
			*soughtItemCount -= itemGetCount
		}
		return // all queue items return: done
	}

	// first part of queue: *soughtItemCount items
	allItems = append(allItems, n.outputQueue[:soughtIC]...)
	copy(n.outputQueue, n.outputQueue[soughtIC:])
	var endIndex = itemGetCount - soughtIC
	for i := endIndex; i < itemGetCount; i++ {
		n.outputQueue[i] = zeroValue
	}
	n.outputQueue = n.outputQueue[:endIndex]
	atomic.AddUint64(&n.unsentCount, uint64(-soughtIC))
	*soughtItemCount = 0

	return
}
