/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/pruntime"
)

// sendThread is the send-thread goroutine function
func (n *NBChan[T]) sendThread(value T) {
	var endCh = *n.threadWait.Load()
	defer close(endCh)
	defer n.sendThreadDeferredClose()
	defer Recover("", nil, n.sendThreadOnError)

	for { // send value loop

		n.sendThreadBlockingSend(value)
		n.updateDataAvailable()

		for { // get value loop

			// check for CloseNow prior to next value
			if n.sendThreadIsCloseNow() {
				return // close now exit
			}

			// if no data, decide on action
			var hasValue bool
			if atomic.LoadUint64(&n.unsentCount) == 0 {

				// for always-threads, wait for alert
				if n.isThreadAlways.Load() && !n.isCloseInvoked.Load() {
					if value, hasValue = n.sendThreadWaitForAlert(); hasValue {
						break // send the value received by alert
					}
					continue // re-check for next action

					// normal thread or after close: exit on no data and no pending sends
				} else if n.sendThreadExitCheck() {
					return // no data, no pending sends: exit thread
				}
			}

			// wait for [NBChan.gets] to be or reach 0
			//	- Get invocations get items before sendThread
			n.getsWait.Wait()

			// try to get value from any queue
			if value, hasValue = n.sendThreadGetNextValue(); hasValue {
				break // send the value fetched from queues
			}

			// there was no data so probably sends are in progress
			//	- wait for any sends to conclude then check for action again
			n.sendsWait.Wait()
		}

		// only send if closeNow has not been invoked
		if !n.sendThreadNewSendCheck() {
			return // CloseNow was invoked: item has been discarded, now exit thread
		}
	}
}

// sendThreadDeferredClose closes the channel if Close was invoked while thread running
//   - invoked by sendThread
func (n *NBChan[T]) sendThreadDeferredClose() {
	if !n.isCloseInvoked.Load() || // - deferred close if isCloseInvoked has become true
		n.isCloseNowInvoked.Load() { // no deferred close for CloseNow
		n.updateDataAvailable()
		return // no deferred close pending return: noop
	}

	// execute deferred close
	n.close() // error is stored in error container. isClosed is active
	// close data waiter
	n.setDataAvailableAfterClose()
}

// sendThreadOnError filters closed channel errors
func (n *NBChan[T]) sendThreadOnError(err error) {
	if pruntime.IsSendOnClosedChannel(err) && n.isCloseNowInvoked.Load() {
		return // ignore if the channel was or became closed
	}
	n.AddError(err)
}

// threadSend sends value on the channel then closes [NBCHan.closesOnThreadSend]
//   - invoked by sendThread holding inputLock
func (n *NBChan[T]) sendThreadBlockingSend(value T) {
	defer close(*n.closesOnThreadSend.Load())
	// count the item just sent — even if panic
	defer atomic.AddUint64(&n.unsentCount, ^uint64(0))

	n.closableChan.Ch() <- value // may block or panic
}

// sendThreadGetNextValue gets the next value for thread
//   - invoked by [NBChan.sendThread]
func (n *NBChan[T]) sendThreadGetNextValue() (value T, hasValue bool) {
	if atomic.LoadUint64(&n.gets) > 0 || atomic.LoadUint64(&n.unsentCount) == 0 {
		return // send thread suspended by Get return: hasValue: false
	}
	n.outputLock.Lock()
	defer n.outputLock.Unlock()

	if hasValue = len(n.outputQueue) > 0 || n.swapQueues(); !hasValue {
		return // no value available return: hasValue false
	}
	value = n.outputQueue[0]
	n.outputQueue = n.outputQueue[1:]
	return // have item return: value: valid, hasValue: true
}

// for always-thread, wait for alert
//   - not if didClose
//   - not if data available
//   - may receive data item
func (n *NBChan[T]) sendThreadWaitForAlert() (value T, hasValue bool) {
	if !n.sendThreadTryAlertHold() {
		return
	}

	// blocks here
	n.threadAlertWait.Wait()

	if valuep := n.threadAlertValue.Load(); valuep != nil {
		n.threadAlertValue.Store(nil)
		hasValue = true
		value = *valuep
	}

	return
}

func (n *NBChan[T]) sendThreadTryAlertHold() (isHold bool) {
	n.threadLock.Lock()
	defer n.threadLock.Unlock()

	if isHold = !n.isCloseInvoked.Load(); !isHold {
		return // no hold
	}
	n.threadAlertWait.HoldWaiters()
	n.threadAlertPending.Store(true)
	return
}

// sendThreadExitCheck stops thread if inside threadLock, unsentCount is 0
//   - may receive a value via alert channel
func (n *NBChan[T]) sendThreadExitCheck() (doStop bool) {
	n.threadLock.Lock()
	defer n.threadLock.Unlock()

	if atomic.LoadUint64(&n.unsentCount) > 0 {
		return // no exit while data available
	} else if atomic.LoadUint64(&n.sends) > 0 {
		return // no exit while sends in progress
	}
	doStop = true
	n.isRunningThread.Store(false)

	return
}

// sendThreadNewSendCheck ensures a new channel send does not start after
// CloseNow
func (n *NBChan[T]) sendThreadNewSendCheck() (doSend bool) {
	n.threadLock.Lock()
	defer n.threadLock.Unlock()

	if doSend = !n.isCloseNowInvoked.Load(); !doSend {
		atomic.AddUint64(&n.unsentCount, ^uint64(0)) // drop the value
		return                                       // CloseNow inoked return: doSend: false
	}

	// arm the read channel
	var ch = make(chan struct{})
	n.closesOnThreadSend.Store(&ch)

	return // doSend: true
}

func (n *NBChan[T]) sendThreadIsCloseNow() (isExit bool) {
	if !n.isCloseNowInvoked.Load() {
		return
	}
	n.threadLock.Lock()
	defer n.threadLock.Unlock()

	if isExit = n.isCloseNowInvoked.Load(); !isExit {
		return
	}
	n.isRunningThread.Store(false)
	return // close now exit
}
