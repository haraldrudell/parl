/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "math"

// Get returns a slice of elementCount or default or zero for all available items held by the channel.
//   - if channel is empty, 0 items are returned
//   - Get is non-blocking
//   - n > 0: max this many items
//   - n == 0 (or <0): all items
//   - Get is panic-free non-blocking error-free thread-safe
func (n *NBChan[T]) Get(elementCount ...int) (allItems []T) {

	// empty NBChan: noop return
	if n.unsentCount.Load() == 0 {
		return // no items available return: nil slice
	}

	// no Get after CloseNow
	if n.isCloseNow.IsInvoked() {
		return
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
		if n := n.unsentCount.Load(); n > 0 {
			allItems = make([]T, 0, n) // approximate size
		}
	}

	n.outputLock.Lock()
	defer n.postGet()

	// get possible item from send thread
	//	- thread decrements unsent count
	if item, itemValid := n.collectSendThreadValue(); itemValid {
		allItems = append(allItems, item)
		if !isAllItems {
			if soughtItemCount--; soughtItemCount == 0 {
				return // fetch complete return
			}
		}
	}

	// fetch from n.outputQueue
	//	- updates unsent count
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

// preGet registers a pending Get invocation prior to outputLock
//   - increases gets and may hold getsWait
//   - block concurrent always-alert
func (n *NBChan[T]) preGet() {
	if n.gets.Add(1) == 1 {
		n.getsWait.HoldWaiters()
		if !n.isThreadAlways.Load() {
			return
		}
		// awayt any Send SendMany always-alert operation has ended
		// and will not be started again before all Get have exited
		n.collectorLock.Lock()
		defer n.collectorLock.Unlock()
	}
}

// postGet is the deferred ending function for [NBChan.Get]
//   - release outputLock
//   - update dataWaitCh
//   - decrease number of Get invocations
//   - if more Get invocations are pending, do nothing
//   - otherwise, release getsWait
//   - check for deferred progress, if so ensure thread progress
func (n *NBChan[T]) postGet() {
	n.outputLock.Unlock()

	// update dataAvailable
	var unsentCount = n.unsentCount.Load()
	n.setDataAvailable(unsentCount > 0)

	// check for last Get
	if n.gets.Add(math.MaxUint64) > 0 {
		return // more Get pending
	}
	n.getsWait.ReleaseWaiters()
	if n.tcIsDeferredProgress() {
		n.tcEnsureThreadProgress()
	}
}

// collectSendThreadValue receives any value in sendThread channel send
//   - invoked by [NBChan.Get] while holding output lock
//   - must await any thread value to ensure values provided in order
//   - thread receives value from:
//   - — Send SendMany that launches thread, but only when sent count 0
//   - — always: thread alert
//   - —on-demand: GetNextValue
func (n *NBChan[T]) collectSendThreadValue() (value T, hasValue bool) {

	// if thread is not running, it does not hold data
	if !n.tcRunningThread.Load() {
		return // thread not running
	}

	// wait for a held state or thread exit
	//	- because this thread holds outputLock,
	//		thread cannot collect additional values
	//	- the only location send-thread can hold with value is
	//		NBChanSendBlock
	var chanState NBChanTState
	select {
	// thread exited
	case <-n.tcThreadExitAwaitable.Ch():
		return // thread exited return
	case chanState = <-n.stateCh():
	}
	// thread held somewhere

	// if it is not send value block, ignore
	//	- NBChanSendBlock is the only wait where thread has value
	if chanState != NBChanSendBlock {
		return // thread is not held in send value
	}

	// request channel send on send block ending
	if !n.tcRequestCollect() {
		return
	}

	// attempt to fetch value from thread
	select {
	case value, hasValue = <-n.closableChan.Ch():
	case <-n.collectChan.Get(1):
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
	n.outputCapacity.Store(uint64(cap(n.outputQueue)))
	n.inputQueue = n.outputQueue0
	n.inputCapacity.Store(uint64(cap(n.inputQueue)))
	n.outputQueue0 = n.outputQueue[:0]
	return
}

// fetchFromOutput gets items from [NBChan.outputQueue]
//   - [NBChan.outputLock] must be held
//   - decrements unsent count
func (n *NBChan[T]) fetchFromOutput(soughtItemCount *int, isAllItems bool, allItems0 []T) (allItems []T) {
	allItems = allItems0

	// empty queue case: no items
	var itemGetCount = len(n.outputQueue)
	if itemGetCount == 0 {
		return // no available items return
	}
	var zeroValue T
	var soughtIC = *soughtItemCount

	// entire queue case: itemCount items
	if isAllItems || itemGetCount <= soughtIC {
		allItems = append(allItems, n.outputQueue...)
		for i := 0; i < itemGetCount; i++ {
			n.outputQueue[i] = zeroValue
		}
		n.outputQueue = n.outputQueue[:0]
		n.unsentCount.Add(uint64(-itemGetCount))
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
	n.unsentCount.Add(uint64(-soughtIC))
	*soughtItemCount = 0

	return
}
