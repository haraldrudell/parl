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
	if n.isCloseInvoked.Load() {
		return // no send after Close(), atomic performance: noop
	}
	n.preSend()
	n.inputLock.Lock()
	defer n.postSend()

	// if Close or CloseNow was invoked, items are discarded
	if n.isCloseInvoked.Load() {
		return // no send after Close() return: noop
	}

	// try providing value to thread
	//	- ensures a thread is running if configured
	//	- updates threadProgressRequired
	if n.alertOrLaunchThreadWithValue(value) {
		return // value was provided to a thread
	}

	// save value in [NBChan.inputQueue]
	if n.inputQueue == nil {
		n.inputQueue = n.newQueue(1) // will allocation proper size
	}
	n.inputQueue = append(n.inputQueue, value)
	n.inputCapacity.Store(uint64(cap(n.inputQueue)))
	n.unsentCount.Add(1)
}

// Send sends many values non-blocking, thread-safe, panic-free and error-free on the channel
//   - if values is length 0 or nil, SendMany only returns count and capacity
func (n *NBChan[T]) SendMany(values []T) {
	var valueCount = len(values)
	if n.isCloseInvoked.Load() || valueCount == 0 {
		return // no send after Close(), atomic performance: noop
	}
	n.preSend()
	n.inputLock.Lock()
	defer n.postSend()

	if n.isCloseInvoked.Load() {
		return // no send after Close() return: noop
	}

	if n.alertOrLaunchThreadWithValue(values[0]) {
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
	n.unsentCount.Add(uint64(valueCount))
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

	// decrement sends
	var sends = int(n.sends.Add(math.MaxUint64))
	if sends == 0 {
		n.sendsWait.ReleaseWaiters()
	}

	// update dataWaitCh
	var unsentCount = n.unsentCount.Load()
	n.setDataAvailable(unsentCount > 0)

	// handle thread progress
	if !n.threadProgressRequired.Load() ||
		sends > 0 ||
		n.gets.Load() > 0 {
		return // thread progress not required or by later thread
	}

	n.tcEnsureThreadProgress()
}

func (n *NBChan[T]) ensureInput(size int) (queue []T) {
	n.inputLock.Lock()
	defer n.inputLock.Unlock()

	if n.inputQueue != nil {
		return
	}
	n.inputQueue = n.newQueue(size)
	return
}

func (n *NBChan[T]) ensureOutput(size int) (queue []T) {
	n.outputLock.Lock()
	defer n.outputLock.Unlock()

	if n.outputQueue != nil {
		return
	}
	n.outputQueue = n.newQueue(size)
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

// alertOrLaunchThreadWithValue attempts to launch thread providing it the value item
//   - Invoked by [NBChan.Send] and [NBChan.SendMany]
//     while holding [NBChan.inputLock]
//   - only invoked when on-demand or always-on threading configured
//   - value has not been added to unsentCount yet
//   - isGetRace indicates pending Get while NBChan is empty
//   - — only matters if didProvideValue is false
//   - caller must hold inputLock
func (n *NBChan[T]) alertOrLaunchThreadWithValue(value T) (didProvideValue bool) {

	//	- NBChan may be configured for no-thread on-demand-thread or always-on thread
	//	- no thread may be running
	//	- this is the only Send SendMany invocation
	//	- Get invocations may be ongoing

	// if NBChan is configured for no thread, value cannot be provided
	if n.noThread.Load() {
		return // no-thread configuration
	}

	// if unsent count is not zero, no change is required to threading
	//	- only Send SendMany which use inputLock can increase this value
	if n.unsentCount.Load() > 0 {
		return // channel is not empty return
	}

	//	- unsent count is zero
	//	- new Get invocations will return prior to outputLock
	//	- a lingering Get may be in progress or waiting for inputLock
	//	- configuration is on-demand-thread or always-thread
	//	- unsent count is zero
	//	- this is only Send/SendMany and there is no Get
	//	- there may be no thread running

	// if Get is in progress, thread should launch later
	if n.tcDeferredProgressCheck() {
		return
	}

	// starting the thread, that is most important
	//	- will go to send value
	if n.tcStartThreadWinner() {
		n.unsentCount.Add(1)
		var hasValue = true
		n.tcStartThread(value, hasValue)
		didProvideValue = true
		n.threadProgressRequired.CompareAndSwap(true, false)
		return // thread was started with value
	}

	// try alerting a waiting always-on thread without waiting
	//	- if successful, thread will go to send value
	if n.isThreadAlways.Load() {
		if didProvideValue = n.tcAlertThread(&value); didProvideValue {
			n.threadProgressRequired.CompareAndSwap(true, false)
			return // always-on thread alerted with value
		}
	}

	// there is a thread running with no items
	//	- on-demand or always
	//	- may be about to: exit getsWait sendsWait
	//	- Close or CloseNow may have been invoked
	//	- if it’s thread exit and unsent items exist, that is problem
	//	- flag it to be dealt with once all Send SendMany Get completes
	n.threadProgressRequired.CompareAndSwap(false, true)

	return
}
