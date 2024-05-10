/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "github.com/haraldrudell/parl/perrors"

// tcStartThread launches the send thread
//   - on Send/SendMany when unsentCount was 0 and no gets in progress
//   - on postGet if unsentCount > 0 and gets in progress went to 0
//   - isRunningThread.CompareAndSwap winner invokes tcStartThread
//   - thread runs until unsentCount is 0 inside threadLock
func (n *NBChan[T]) tcStartThread(value T, hasValue bool) { go n.sendThread(value, hasValue) }

// tcStartThreadWinner attempts to be the thread that gets to launch sendThread
//   - isWinner true: tcStartThread should be invoked by this invocation
//     this thread won to start a thread
//   - isWinner false: a thread is running or being started
//   - inside threadLock
//   - on isRunningThread true, closesOnThreadSend must be valid
func (n *NBChan[T]) tcStartThreadWinner() (isWinner bool) {
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	// no thread launch after Close CloseNow invoked
	if n.isCloseInvoked.Load() {
		return
	}

	// note that thread did at one time launch
	n.tcDidLaunchThread.CompareAndSwap(false, true)

	// check for thread already running or this invocation should not launch it
	if isWinner = n.tcRunningThread.CompareAndSwap(false, true); !isWinner {
		return // thread was already running return
	}
	// arm thread exit awaitable
	n.tcThreadExitAwaitable.Open()

	return
}

// tcAlertOrLaunchThreadWithValue attempts to launch thread providing it the value item
//   - didProvideValue true: value was prrovided to thread via launch or alert,
//     unsent count was incremented
//   - Invoked by [NBChan.Send] and [NBChan.SendMany]
//     while holding [NBChan.inputLock]
//   - value has not been added to unsentCount yet
//   - isGetRace indicates pending Get while NBChan is empty
//   - — only matters if didProvideValue is false
//   - caller must hold inputLock to prevent parallel invocations where
//     multiple threads handle unsent count zero
func (n *NBChan[T]) tcAlertOrLaunchThreadWithValue(value T) (didProvideValue bool) {

	// if unsent count is zero, a thread may have to be alerted or launched
	//	- no thread may have been launched
	//	- on-demand thread may have exited on unsent count zero
	//	- always thread may await an alert
	// - no action is necessary if:
	//	- — threading is not used
	//	- — unsent count is not zero
	//	- — Gets are in progress for which a thread would be a detour
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
	n.progressLock.Lock()
	defer n.progressLock.Unlock()

	// if Get is in progress, thread should launch later
	if n.tcDeferredProgressCheck() {
		return
	}

	// starting the thread, that is most important
	//	- a launched thread will go to send value
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

	// there is a thread running without holding an item
	//	- on-demand or always
	//	- may be executing towards: exit getsWait sendsWait alert
	//	- Close or CloseNow may have been invoked
	//	- at end of Send SendMany Get invocations, a static thread-state must be awaited
	//		so that progress is guaranteed
	//	- issue is that:
	//	- — an on-demand thread may exit while items are present
	//	- — an always thread may soon need an alert
	//	- flag it to be dealt with once all Send SendMany Get completes
	n.threadProgressRequired.CompareAndSwap(false, true)

	return
}

// tcSendEnsureProgress handdles thread progress on final Send SendMany conclusion
func (n *NBChan[T]) tcSendEnsureProgress() {
	if !n.threadProgressRequired.Load() {
		return // thread progress not required
	} else if n.tcDeferredProgressCheck() {
		return // deferred to Get conclusion
	}
	n.tcEnsureThreadProgress()
}

// tcEnsureThreadProgress ensures that thread is not exiting
//   - final Send SendMany concluded while no Get in progress or
//   - final Get concluded
//   - invoked when threadProgressRequired observed true
func (n *NBChan[T]) tcEnsureThreadProgress() {

	//	- Send and SendMany encountered unsent count zero
	//	- it is on-demand or always thread
	//	- because a thread could not be started or alerted,
	//		ensuring thread progress was deferred

	// check for no items
	if n.tcZeroProgress() {
		return
	}

	// await thread static state
	select {
	// thread exit
	case <-n.tcThreadExitAwaitable.Ch():
		// NBChanSendBlock NBChanAlert NBChanGets NBChanSends
	case threadState := <-n.stateCh():
		switch threadState {
		case NBChanSendBlock:
			n.tcSendProgress()
		case NBChanAlert:
			// unblock always-thread awaiting alert
			n.tcAlertProgress()
		}
		return // progress attempted or thread holding in Gets Sends
	}
	// thread did exit

	if n.isThreadAlways.Load() {
		// exiting always thread is close or panic
		return // exiting always thread: noop
	}

	// unless unsent count is again zero, the thread must not exit

	// close now
	if n.isCloseNow.IsInvoked() {
		return // out of items return: thread exit ok
	}

	// check close status
	if n.isCloseComplete() {
		return // underlying channel has closed: noop
	}

	// seek permisssion to start thread
	if !n.tcStartProgress() {
		return // some other thread started the send thread
	}

	// start thread without value
	var value T
	var hasValue bool
	n.tcStartThread(value, hasValue)
}

// isCloseComplete checks for underlying channel closed
//   - by close
//   - by thread exit in deferred close
//   - by close now
func (n *NBChan[T]) isCloseComplete() (isClosed bool) {
	n.inputLock.Lock()
	defer n.inputLock.Unlock()

	isClosed = n.closableChan.IsClosed()

	return
}

// tcSendProgress affirms progress by thread in send block
func (n *NBChan[T]) tcSendProgress() {
	n.progressLock.Lock()
	defer n.progressLock.Unlock()

	var threadState NBChanTState
	select {
	case <-n.tcThreadExitAwaitable.Ch():
		return
	case threadState = <-n.stateCh():
	}
	if threadState == NBChanSendBlock {
		n.threadProgressRequired.Store(false)
	}
}

// tcStartProgress affirms progress by obtaining thread-start permission
func (n *NBChan[T]) tcStartProgress() (isProgress bool) {
	n.progressLock.Lock()
	defer n.progressLock.Unlock()

	isProgress = n.tcStartThreadWinner()
	return
}

// tcZeroProgress affirms progress by observing unsent count zero
func (n *NBChan[T]) tcZeroProgress() (isProgress bool) {
	n.progressLock.Lock()
	defer n.progressLock.Unlock()

	if isProgress = n.unsentCount.Load() == 0; isProgress {
		n.threadProgressRequired.Store(false)
	}
	return
}

// tcAlertProgress attempts to alert a thread
//   - if alert succeeds, threadProgressRequired is reset
func (n *NBChan[T]) tcAlertProgress() {
	n.progressLock.Lock()
	defer n.progressLock.Unlock()

	if n.tcAlertNoValue() {
		n.threadProgressRequired.Store(false)
	}
}

// tcAlertNoValue alerts any waiting always-thread
func (n *NBChan[T]) tcAlertNoValue() (didAlert bool) { return n.tcAlertThread(nil) }

// tcAlertThread alerts any waiting always-thread
//   - invoked from Send/SendMany
//   - value has not been added to unsentCount yet
//   - increments unsentCount if value is non-nil and was provided to thread
func (n *NBChan[T]) tcAlertThread(valuep *T) (didAlert bool) {

	// can this thread send?
	if didAlert = n.gets.Load() == 0 && // only when no Get in progress
		// only if send winner
		n.threadChWinner.CompareAndSwap(true, false); !didAlert {
		return // no
	}

	if valuep != nil {
		n.unsentCount.Add(1)
	}
	select {
	case n.threadCh <- valuep:
	case <-n.threadCh2:
	}
	return
}

// tcDeferredProgressCheck may mark threadProgressRequired by Get pending
//   - may be invoked by Send SendMany when unsent count is zero
//   - if Get is in progress, it is not efficient to provide items to thread
//   - the lock ensures that if Send SendMany detects Get in progress,
//     on Get conclusion, the last Get is guaranteed to check threadProgressRequired
//   - invoked while holding progressLock
func (n *NBChan[T]) tcDeferredProgressCheck() (isDeferred bool) {
	n.getProgressLock.Lock()
	defer n.getProgressLock.Unlock()

	if isDeferred = n.gets.Load() > 0; !isDeferred {
		return
	}
	n.threadProgressRequired.Store(true)
	return
}

func (n *NBChan[T]) tcIsDeferredProgress() (needProgress bool) {
	n.getProgressLock.Lock()
	defer n.getProgressLock.Unlock()

	needProgress = n.threadProgressRequired.Load()

	return
}

// request from thread to send value on end of send block
//   - lock to ensure thread does not exit during preparations
//   - only succeeds if a thread is in send block
func (n *NBChan[T]) tcRequestCollect() (collectIsActive bool) {
	n.collectLock.Lock()
	defer n.collectLock.Unlock()

	if !n.tcSendBlock.Load() {
		return
	}

	var collectChan = n.collectChan.Get(1)
	select {
	case <-collectChan:
	default:
	}
	if len(collectChan) > 0 {
		panic(perrors.NewPF("collectChanreadfailed"))
	}

	n.tcCollectRequest.Store(true)
	collectIsActive = true
	return
}

// returns a channel producing values if thread is holding:
//   - NBChanSendBlock NBChanAlert NBChanGets NBChanSends
func (n *NBChan[T]) stateCh() (ch chan NBChanTState) {
	if chp := n.tcState.Load(); chp != nil {
		ch = *chp
		return
	}
	ch = make(chan NBChanTState)
	if n.tcState.CompareAndSwap(nil, &ch) {
		return
	}
	ch = *n.tcState.Load()
	return
}
