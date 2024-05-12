/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "math"

// Send sends a single value on the channel
//   - non-blocking, thread-safe, panic-free and error-free
//   - if Close or CloseNow was invoked, items are discarded
func (n *NBChan[T]) Send(value T) {
	if n.isCloseInvoked.IsInvoked() {
		return // no send after Close(), atomic performance: noop
	}
	n.preSend()
	n.inputLock.Lock()
	defer n.postSend()

	// if Close or CloseNow was invoked, items are discarded
	if n.isCloseInvoked.IsInvoked() {
		return // no send after Close() return: noop
	}

	// try providing value to thread
	//	- ensures a thread is running if configured
	//	- updates threadProgressRequired
	if n.tcAlertOrLaunchThreadWithValue(value) {
		return // value was provided to a thread
	}

	// save value in [NBChan.inputQueue]
	if n.inputQueue == nil {
		n.inputQueue = n.newQueue(1) // will allocation proper size
	}
	n.inputQueue = append(n.inputQueue, value)
	n.inputCapacity.Store(uint64(cap(n.inputQueue)))
	n.tcAddProgress(1)
}

// Send sends many values non-blocking, thread-safe, panic-free and error-free on the channel
//   - if values is length 0 or nil, SendMany only returns count and capacity
func (n *NBChan[T]) SendMany(values []T) {
	var valueCount = len(values)
	if n.isCloseInvoked.IsInvoked() || valueCount == 0 {
		return // no send after Close(), atomic performance: noop
	}
	n.preSend()
	n.inputLock.Lock()
	defer n.postSend()

	if n.isCloseInvoked.IsInvoked() {
		return // no send after Close() return: noop
	}

	if n.tcAlertOrLaunchThreadWithValue(values[0]) {
		values = values[1:]
		valueCount--
		if valueCount == 0 {
			return // one value handed to thread return: complete
		}
	}

	// save values in [NBChan.inputQueue]
	if n.inputQueue == nil {
		n.inputQueue = n.newQueue(valueCount)
	}
	n.inputQueue = append(n.inputQueue, values...)
	n.inputCapacity.Store(uint64(cap(n.inputQueue)))
	n.tcAddProgress(valueCount)
}

// preSend registers a Send or SendMany invocation pre-inputLock
//   - send count is in [NBChan.sends]
//   - handles [NBChan.sendsWait] that prevents a thread from exiting
//     during Send SendMany invocations
func (n *NBChan[T]) preSend() {
	if n.sends.Add(1) == 1 {
		n.sendsWait.HoldWaiters()
	}
}

// post send is deferred for [NBChan.Send] and [NBChan.SendMany]
//   - release inputLock
//   - alert thread if no pending Get and values ar present
func (n *NBChan[T]) postSend() {
	n.inputLock.Unlock()

	// update dataWaitCh
	n.updateDataAvailable()

	// decrement sends
	if n.sends.Add(math.MaxUint64) == 0 {
		n.sendsWait.ReleaseWaiters()
	}

	// ensure progress
	for {
		if isZeroObserved, isGets := n.tcIsDeferredSend(); !isZeroObserved || isGets {
			// progress not required or
			// deferred by Get invocations
			return
		} else if !n.tcAwaitProgress() {
			// progress was secured
			return
		} else if n.gets.Load() > 0 {
			// subsequent Send SendMany exist
			//	- after sends decrement, those will arrive at ensure progress
			return
		}
	}
}

// ensureInput allocates or enlarges for [NBChan.SetAllocationSize]
func (n *NBChan[T]) ensureInput(size int) (queue []T) {
	n.inputLock.Lock()
	defer n.inputLock.Unlock()

	if n.inputQueue != nil {
		return
	}
	n.inputQueue = n.newQueue(size)
	return
}

// newQueue allocates a new queue slice
//   - capacity is at least count elements
//   - the slice is empty
func (n *NBChan[T]) newQueue(count int) (queue []T) {

	// determine size
	var size = int(n.allocationSize.Load())
	if size > 0 {
		if count > size {
			size = count
		}
	} else {
		size = defaultNBChanSize
		if count > size {
			size = count * 2
		}
	}

	// return allocated zero-length queue
	return make([]T, size)[:0]
}
