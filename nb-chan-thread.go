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
	var zeroValue T
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
			hasValue = false
			value = zeroValue
		}

		// obtain next value loop
		for {

			// check for CloseNow prior to next value
			if n.sendThreadIsCloseNow() {
				return // close now exit: immediate discard and exit
			}

			// if no data, decide on action
			if n.unsentCount.Load() == 0 {
				n.sendThreadZero()

				// always-thread not in deferred close: wait for alert
				if n.isThreadAlways.Load() && !n.isCloseInvoked.IsInvoked() {
					// blocks here
					if value, hasValue = n.sendThreadWaitForAlert(); hasValue {
						break // send the value received by alert
					}
					continue // re-check closeNow and unsent count for next action

					// on-demand thread or always in deferred close:
					// exit on no data and no pending sends
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
					case <-ch: // Get ceased
					case n.stateCh() <- NBChanGets: // respond is in Gets wait
						continue
					}
					break
				}
			}

			// try to get value from any queue
			if value, hasValue = n.sendThreadGetNextValue(); hasValue {
				break // send the value fetched from queues
			}

			// unsent count has reached zero or Get is in progress
			//	- wait for any sends to conclude that may provide additional items
			if ch := n.sendsWait.Ch(); ch != nil {
				for {
					select {
					case <-ch:
					case n.stateCh() <- NBChanSends:
						continue
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

// sendThreadDeferredClose may close the underlying channel
//   - is how sendThread executes deferred close
//   - closes if Close was invoked while thread running and not CloseNow
//   - invoked by sendThread on exit
//   - updates dataWaitCh
func (n *NBChan[T]) sendThreadDeferredClose() {

	// is it deferred close?
	if !n.isCloseInvoked.IsInvoked() || // no: Close has not been invoked
		n.isCloseNow.IsInvoked() { // CloseNow overrides deferred close
		n.updateDataAvailable()
		return // no deferred close pending return: noop
	}

	// for on-demand thread, ensure out of data
	if !n.isThreadAlways.Load() {

		if n.unsentCount.Load() > 0 {
			return
		}
		// tcThread
	}

	// execute deferred close
	//	- error is stored in error container. isClosed is active
	n.executeChClose()
	// close data waiter
	n.setDataAvailableAfterClose()
}

// sendThreadZero notifies background that thread
// took action on unsent count zero
func (n *NBChan[T]) sendThreadZero() {
	n.tcDoProgressRaised(true)
	n.tcProgressLock.Lock()
	defer n.tcProgressLock.Unlock()

	if n.unsentCount.Load() == 0 {
		n.tcProgressRequired.Store(true)
	}
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
	defer n.updateDataAvailable()
	// count the item just sent — even if panic
	defer n.unsentCount.Add(math.MaxUint64)
	// clear two-chan receive second channel
	n.collectChanActive.Store(nil)
	// receive value with default has proven to result in default. Therefore:
	//	- two-chan receive is used by tcCollectThreadValue to prevent deadlock and aba
	//	- send-thread provides an atomic true and a nil atomic channel value
	//		upon commencing send operation
	//	- the atomic true allows other threads to write the atomic channel value and
	//		reset the atomic true to false
	//	- a winner thread observing the atomic true value stores a 1-size empty channel,
	//		and proceeds if it is able to set the atomic true value to false
	//	- at end of send operation, send-thread attempts to change the atomic value from true to false
	//	- if the atomic value was true, no two-chan receive is in progress
	//	- otherwise, send-thread sends on the atomic channel
	//	- thereby, send-thread will not enter dead-lock and avoids aba-issue
	defer n.sendThreadBlockingSendEnd()
	// tcSendBlock makes collectChanActive available to tcCollectThreadValue threads
	n.tcSendBlock.Store(true)

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

// sendThreadBlockingSendEnd completes any two-chan receive operation
func (n *NBChan[T]) sendThreadBlockingSendEnd() {
	// check if two-chan send was initiated
	if n.tcSendBlock.CompareAndSwap(true, false) {
		return // no value collect
	}

	// send to ensure tcCollectThreadValue is not blocked
	if cp := n.collectChanActive.Load(); cp != nil {
		*cp <- struct{}{}
	}
}

// sendThreadGetNextValue gets the next value for thread
//   - invoked by [NBChan.sendThread]
//   - fails if pending Get or unsentCount ends up zero
func (n *NBChan[T]) sendThreadGetNextValue() (value T, hasValue bool) {
	if n.gets.Load() > 0 || n.unsentCount.Load() == 0 {
		return // send thread suspended by Get return: hasValue: false
	}
	// if a thread holding outputLock awaited thread state,
	// acquiring outputLock here could cause dead-lock
	//	- only Get invocations do this
	//	- therefore, ensure outputLock is not acquired while Get
	//		in progress
	n.collectorLock.Lock()
	defer n.collectorLock.Unlock()

	if n.gets.Load() > 0 {
		return // cancel: Get in progress
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

// sendThreadWaitForAlert allows an always-on thread to await alert
//   - the alert is a two-chan send that may provide a value
//   - always threads do not exit, instead at end of data
//     they wait for background events:
//   - not if didClose
//   - not if data available
//   - an alert that may provide a data item
func (n *NBChan[T]) sendThreadWaitForAlert() (value T, hasValue bool) {

	// reset atomic channel
	n.alertChan2Active.Store(nil)
	// n.threadChWinner true exposes channels to clients
	n.tcAlertActive.Store(true)
	// sending on threadCh2 ensures no client is hanging
	defer n.sendThreadAlertEnd()

	// blocks here
	//	- n.threadCh must be unbuffered for effect to be immediate
	//	- n.threadCh2 is present to prevent client from hanging in threadCh send
	for {
		select {
		// wait for alert
		case valuep := <-n.alertChan.Get():
			if hasValue = valuep != nil; hasValue {
				value = *valuep
				n.unsentCount.Add(1)
			}
			return
			// broadcast Alert wait
		case n.stateCh() <- NBChanAlert:
		}
	}
}

func (n *NBChan[T]) sendThreadAlertEnd() {
	// see if two-chan send operation in progress
	if n.tcAlertActive.CompareAndSwap(true, false) {
		return // no
	} else if cp := n.alertChan2Active.Load(); cp != nil {
		*cp <- struct{}{}
	}
}

// sendThreadExitCheck stops thread if inside threadLock, unsentCount is 0
//   - doStop true: channel has been read to end and no Send SendMany active
//   - invoked when thread has detected Close invocation and is in deferred close
func (n *NBChan[T]) sendThreadExitCheck() (doStop bool) {
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	if doStop = //
		n.unsentCount.Load() == 0 && // only stop if out of data
			n.sends.Load() == 0; // while no sends in progress
	!doStop {
		return
	}

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

// sendThreadIsCloseNow checks for CloseNow invocation
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
