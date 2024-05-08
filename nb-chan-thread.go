/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"math"

	"github.com/haraldrudell/parl/pruntime"
)

// sendThread feeds values to the send channel
//   - may be on-demand or always-on thread
//   - verbose='NBChan.*sendThread'
func (n *NBChan[T]) sendThread(value T, hasValue bool) {
	// signal thread exit to CloseNow if it is waiting
	defer n.tcThreadExitAwaitable.Close()
	// execute possible deferred close from Close invocation
	defer n.sendThreadDeferredClose()
	defer Recover(func() DA { return A() }, nil, n.sendThreadOnError)

	// send value loop
	for {

		if hasValue {
			// send the value: blocks here
			//	- until consumer receive, Get, CloseNow or panic
			//	- decrements unsent count
			n.sendThreadBlockingSend(value)
			n.updateDataAvailable()
			hasValue = false
		}

		// obtain next value loop
		for {

			// check for CloseNow prior to next value
			if n.sendThreadIsCloseNow() {
				return // close now exit: immediate discard and exit
			}

			// if no data, decide on action
			if n.unsentCount.Load() == 0 {

				// for always-threads, wait for alert
				if n.isThreadAlways.Load() && !n.isCloseInvoked.Load() {
					// blocks here
					if value, hasValue = n.sendThreadWaitForAlert(); hasValue {
						break // send the value received by alert
					}
					continue // re-check for next action

					// normal thread or after close: exit on no data and no pending sends
				} else if n.sendThreadExitCheck() {
					// on-demand thread or always-on after Close exits here
					return // no data, no pending sends: exit thread
				}
			} // obtain next value loop
			// there is more data to send

			if hasValue {
				break // send the always-on thread value from alert
			}
			// is on-demand thread and data is available

			// wait for [NBChan.gets] to be or reach 0
			//	- Get invocations get items before sendThread
			if ch := n.getsWait.Ch(); ch != nil {
				for {
					select {
					case <-ch:
					case n.stateCh() <- NBChanGets:
					}
					break
				}
			}

			// try to get value from any queue
			if value, hasValue = n.sendThreadGetNextValue(); hasValue {
				break // send the value fetched from queues
			}

			// there was no data so probably sends are in progress
			//	- wait for any sends to conclude then check for action again
			if ch := n.sendsWait.Ch(); ch != nil {
				for {
					select {
					case <-ch:
					case n.stateCh() <- NBChanSends:
					}
					break
				}
			}
		}
		// a value was obtained

		// only send if closeNow has not been invoked
		if !n.sendThreadNewSendCheck() {
			return // CloseNow: item is discarded, thread exits
		}
	}
}

// sendThreadDeferredClose may closes the underlying channel
//   - is how sendThread executes deferred close
//   - closes if Close was invoked while thread running and not CloseNow
//   - invoked by sendThread on exit
//   - updates dataWaitCh
func (n *NBChan[T]) sendThreadDeferredClose() {
	if !n.isCloseInvoked.Load() || // - deferred close if isCloseInvoked has become true
		n.isCloseNow.IsInvoked() { // no deferred close for CloseNow
		n.updateDataAvailable()
		return // no deferred close pending return: noop
	}

	// execute deferred close
	//	- // error is stored in error container. isClosed is active
	n.executeChClose()
	// close data waiter
	n.setDataAvailableAfterClose()
}

// sendThreadOnError submits thread panic
//   - ignores send on closed channel after closenow
func (n *NBChan[T]) sendThreadOnError(err error) {
	if pruntime.IsSendOnClosedChannel(err) && n.isCloseNow.IsInvoked() {
		return // ignore if the channel was or became closed
	}
	n.AddError(err)
}

// sendThreadBlockingSend sends blocking on consumer-receive channel
//   - decrements unsent count
//   - blocks until:
//   - — consumer read the value
//   - — Get empties the channel using collectSendThreadValue
//   - — CloseNow discards the value using discardSendThreadValue
//   - invoked by sendThread holding inputLock
func (n *NBChan[T]) sendThreadBlockingSend(value T) {
	// count the item just sent — even if panic
	defer n.unsentCount.Add(math.MaxUint64)
	// receive value with default has proven to result in default. Therefore:
	//	- background may request value sent on collectChan at end of send block
	//	- background can then receive from both channels which is gauranteed to end
	//	- there can only be one thread and only one background Get may execute at any time
	defer n.sendThreadBlockingSendEnd()
	// reset collect request
	n.tcCollectRequest.CompareAndSwap(true, false)
	// indicate thread in send block
	n.tcSendBlock.Store(true)
	defer n.tcSendBlock.Store(false)

	for {
		select {
		// send value to consumer or Get
		//	- may block or panic
		case n.closableChan.Ch() <- value:
			return
		case n.stateCh() <- NBChanSendBlock:
		}
	}
}

func (n *NBChan[T]) sendThreadBlockingSendEnd() {
	n.collectLock.Lock()
	defer n.collectLock.Unlock()

	// was there a collect request during send block?
	if !n.tcCollectRequest.CompareAndSwap(true, false) {
		return
	}

	// send
	var collectChan = n.collectChan.Get(1)
	collectChan <- struct{}{}
}

// sendThreadGetNextValue gets the next value for thread
//   - invoked by [NBChan.sendThread]
//   - fails if pending Get or unsentCount ends up zero
func (n *NBChan[T]) sendThreadGetNextValue() (value T, hasValue bool) {
	if n.gets.Load() > 0 || n.unsentCount.Load() == 0 {
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

// sendThreadWaitForAlert allows an always-on thread to wait for an alert
//   - always threads do not exit, instead at end of data
//     they wait for background events:
//   - not if didClose
//   - not if data available
//   - may receive data item
func (n *NBChan[T]) sendThreadWaitForAlert() (value T, hasValue bool) {

	// arm alert waiter
	n.threadAlertWait.HoldWaiters()
	var awaitableCh = n.threadAlertWait.Ch()
	if awaitableCh == nil {
		return // not waiting
	}

	// blocks here
	for {
		select {
		// wait for ReleaseWaiters
		case <-awaitableCh:
			// broadcast Alert wait
		case n.stateCh() <- NBChanAlert:
			continue
		}
		break
	}

	// collect possible value
	if valuep := n.threadAlertValue.Load(); valuep != nil {
		n.threadAlertValue.Store(nil)
		hasValue = true
		value = *valuep
	}

	return
}

// sendThreadExitCheck stops thread if inside threadLock, unsentCount is 0
//   - doStop true: channel has been read to end and no Send SendMany active
//   - invoked when thread has detected Close invocation and is in deferred close
func (n *NBChan[T]) sendThreadExitCheck() (doStop bool) {
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	if n.unsentCount.Load() > 0 {
		return // no exit while data available
	} else if n.sends.Load() > 0 {
		return // no exit while sends in progress
	}
	doStop = true
	n.tcRunningThread.Store(false)

	return
}

// sendThreadNewSendCheck ensures a new channel send does not start after
// CloseNow
//   - on closeNow, value is discarded
//   - re-arms closesOnThreadSend
func (n *NBChan[T]) sendThreadNewSendCheck() (doSend bool) {

	if doSend = !n.isCloseNow.IsInvoked(); !doSend {
		n.unsentCount.Add(math.MaxUint64) // drop the value
		return                            // CloseNow inoked return: doSend: false
	}

	return // doSend: true
}

// sendThreadIsCloseNow checks for CloseNow
//   - isExit true: CloseNow was invoked
func (n *NBChan[T]) sendThreadIsCloseNow() (isExit bool) {
	if !n.isCloseNow.IsInvoked() {
		return // no CloseNow invocation return
	}
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	if isExit = n.isCloseNow.IsInvoked(); !isExit {
		return
	}
	n.tcRunningThread.Store(false)
	return // close now exit
}
