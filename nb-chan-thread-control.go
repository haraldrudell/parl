/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

const (
	// tcStartThread
	NoValue = false
	// tcStartThread
	HasValue = true
)

// tcCreateWinner attempts to be the thread that gets to launch sendThread
//   - isWinner true: tcStartThread should be invoked by this invocation
//     this thread won to start a thread
//   - isWinner false: a thread is running or being started
//   - inside threadLock
//   - on isRunningThread true, closesOnThreadSend must be valid
func (n *NBChan[T]) tcCreateWinner() (isWinner bool) {
	// lock creating critical section with setting isCloseInvoked to true
	//	- once close closeNow set isCloseInvoked to true,
	//		it is certain no thread will be permitted to launch
	//	- atomizes isClose false detection with setting tcRunningThread
	//		to true
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	// no thread launch after Close CloseNow invoked
	if n.isCloseInvoked.IsInvoked() {
		return
	}

	// note that thread did at one time launch
	//	- before tcRunningThread is set to true
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
	//	- configuration is on-demand-thread or always-thread
	//	- there may be no thread running
	//	- progress must now be guaranteed by:
	//	- — launching the send-thread
	//	- — alerting an always-thread
	//	- — for on-demand thread in unknown state,
	//		observing it in a state once data has been added to inputBuffer.
	//		It may otherwise exit

	// if Get is in progress, thread should launch later
	if _, isGets := n.tcIsDeferredSend(); isGets {
		return // deferred: threadProgressRequired true
	}

	// starting the thread, that is most important
	//	- a launched thread will go to send value
	if _, didProvideValue = n.tcCreateProgress(); didProvideValue {
		n.unsentCount.Add(1)
		go n.sendThread(value, HasValue)
		return // thread was started with value
	}

	// try alerting a waiting always-on thread without waiting
	//	- if successful, thread will go to send value
	if didProvideValue = n.tcAlertProgress(&value); didProvideValue {
		return // always-on thread alerted with value
	}

	// there is a thread running without holding an item
	//	- on-demand or always
	//	- may be executing towards: exit getsWait sendsWait alert
	//	- Close or CloseNow may have been invoked
	//	- at end of Send SendMany Get invocations, a static thread-state must be awaited
	//		so that progress is guaranteed
	//	- issue is that:
	//	- — an on-demand thread may exit while items are present
	//	- — an always thread may enter alert state where it must be alerted
	//	- flag it to be dealt with once all Send SendMany Get completes
	return
}

// tcCreateProgress seeks progress by creating the send-thread
//   - honorProgressRaised true: isProgress is only true if tcProgressRaised remains false
//   - isProgress: isCreateThread is also true, this operation is thread progress
//   - isCreateThread: should invoke tcStartThread
func (n *NBChan[T]) tcCreateProgress(honorProgressRaised ...bool) (isProgress, isCreateThread bool) {
	n.tcProgressLock.Lock()
	defer n.tcProgressLock.Unlock()

	if isCreateThread = n.tcCreateWinner(); !isCreateThread {
		return // no progress no thread to be created
	} else if len(honorProgressRaised) > 0 && honorProgressRaised[0] && n.tcDoProgressRaised() {
		return // progressRaised true, so this is not progress
	}
	isProgress = true
	return
}

// HonorProgressRaised ignore any progress made if tcProgressRaised is true
const HonorProgressRaised = true

// record progress regardless of tcProgressRaised
const IgnoreProgressRaised = false

// tcAlertProgress attempts progress via alert ignoring tcProgressRaised
func (n *NBChan[T]) tcAlertProgress(valuep ...*T) (isProgress bool) {
	isProgress, _ = n.tcAlertProgress2(IgnoreProgressRaised, valuep...)
	return
}

// tcAlertProgress2 attempts progress via alert optionally honoring tcProgressRaised
//   - honorProgressRaised true: isProgress is only true if tcProgressRaised remains false
//   - valuep: optional value provided with alert
//   - isProgress operation counts as progress
//   - didAlert an alert was successfully sent
func (n *NBChan[T]) tcAlertProgress2(honorProgressRaised bool, valuep ...*T) (isProgress, didAlert bool) {
	n.tcProgressLock.Lock()
	defer n.tcProgressLock.Unlock()

	// try alerting the thread
	var valuep0 *T
	if len(valuep) > 0 {
		valuep0 = valuep[0]
	}
	if didAlert = //
		n.tcAlertActive.Load() && // only when send-thread awaits alert
			n.tcAlertThread(valuep0); //
	!didAlert {
		return // no successful alert
	} else if isProgress = !honorProgressRaised || !n.tcDoProgressRaised(); !isProgress {
		return // was progress raised true
	}
	n.tcProgressRequired.Store(false)
	return
}

func (n *NBChan[T]) tcAddProgress(count int) {
	n.tcDoProgressRaised(true)
	n.tcProgressLock.Lock()
	defer n.tcProgressLock.Unlock()

	if n.unsentCount.Add(uint64(count)) == uint64(count) && !n.noThread.Load() {
		n.tcProgressRequired.Store(true)
	}
}

// tcAwaitProgress awaits a static state from thread then ensures progress
func (n *NBChan[T]) tcAwaitProgress() (threadProgressRequired bool) {
	n.tcAwaitProgressLock.Lock()
	defer n.tcAwaitProgressLock.Unlock()

	// check if progress action remains required
	if threadProgressRequired = n.tcProgressRequired.Load(); !threadProgressRequired {
		return
	}
	// reset progress raised, ie.
	//	- thread observing unsent count zero
	//	- Send SendMany increasing unsent count from zero
	n.tcDoProgressRaised(false)

	// await static thread status
	//	- cannot hold tcProgressLock lock
	var threadState NBChanTState
	select {
	// thread exit
	case <-n.tcThreadExitAwaitable.Ch():
		if !n.isThreadAlways.Load() {
			var isProgress, isCreateThread = n.tcCreateProgress(HonorProgressRaised)
			if isCreateThread {
				// start thread without value
				var value T
				go n.sendThread(value, NoValue)
			}
			threadProgressRequired = !isProgress
		}
		return
	case threadState = <-n.stateCh():
	}
	// received thread status

	// alert
	if threadState == NBChanAlert {
		if isProgress, _ := n.tcAlertProgress2(HonorProgressRaised); isProgress {
			threadProgressRequired = false
		}
		return
	}

	// all other states are good states
	//	- if a zero unsent period occurred in the meantime,
	//	- tcProgressRaised is true and
	//	- the operation was not progress
	threadProgressRequired = n.tcDoProgressRaised()

	return
}

// tcDoProgressRaised updates and or returns tcProgressRaised
//   - tcProgressRaised is used when awaiting thread-state since tcProgressLock
//     cannot be held
//   - when tcProgressRequired is to be reset, this opeation is ignored if
//     tcProgressRaised is true, ie.
//   - — the send-thread took action on unsent count zero or
//   - — Send SendMany added items from unsent count zero
func (n *NBChan[T]) tcDoProgressRaised(isRaised ...bool) (wasRaised bool) {
	if len(isRaised) == 0 {
		wasRaised = n.tcProgressRaised.Load()
		return
	}
	wasRaised = n.tcProgressRaised.Swap(isRaised[0])
	return
}

// tcAlertThread alerts any waiting always-thread
//   - invoked from Send/SendMany
//   - value has not been added to unsentCount yet
//   - increments unsentCount if value is non-nil and was provided to thread
func (n *NBChan[T]) tcAlertThread(valuep ...*T) (didAlert bool) {
	// atomizes always-alert with detecting no Get in progress
	n.collectorLock.Lock()
	defer n.collectorLock.Unlock()

	// filter send request
	if n.gets.Load() > 0 || // not when no Get in progress
		!n.tcAlertActive.Load() { // only when send-thread awaits alert
		return // no alert now
	}

	// prepare two-chan send second channel
	var alertChan2 = n.alertChan2.Get(1)
	if len(alertChan2) > 0 {
		<-alertChan2
	}
	n.alertChan2Active.Store(&alertChan2)

	// verify that two-chan send still available
	if !n.tcAlertActive.CompareAndSwap(true, false) {
		return
	}

	// send value to always-thread
	//	- always thread increments unsentCount on reception
	var valuep0 *T
	if len(valuep) > 0 {
		valuep0 = valuep[0]
	}
	select {
	case n.alertChan.Get() <- valuep0:
		didAlert = true
	case <-alertChan2:
	}
	return
}

// tcIsDeferredSend checks for a progress requirement
//   - invoked by the last ending Get invocation
func (n *NBChan[T]) tcIsDeferredSend() (isProgressRequired, isGets bool) {
	n.tcGetProgressLock.Lock()
	defer n.tcGetProgressLock.Unlock()

	isProgressRequired = n.tcProgressRequired.Load()
	isGets = n.gets.Load() > 0

	return
}

// tcCollectThreadValue receives any value in sendThread channel send
//   - invoked by [NBChan.Get] while holding output lock
//   - must await any thread value to ensure values provided in order
//   - thread receives value from:
//   - — Send SendMany that launches thread, but only when sent count 0
//   - — always: thread alert
//   - —on-demand: GetNextValue
func (n *NBChan[T]) tcCollectThreadValue() (value T, hasValue bool) {

	// if thread is not running, it does not hold data
	if !n.tcRunningThread.Load() {
		return // thread not running
	}

	// await static thread state
	//	- state must be awaited since the thread may be in progress
	//		with a value towards NBChanSendBlock
	//	- if NBChanSendBlock, threads hold a value
	//	- in all other static thread states, thread holds no value
	//	- because this thread holds outputLock,
	//		on-demand thread cannot collect additional values
	//	- collectorLock ensures that no always-thread alerts are carried out
	//		while Get is in progress
	select {
	// thread exited
	case <-n.tcThreadExitAwaitable.Ch():
		return // thread exited return
	case chanState := <-n.stateCh():
		// if it is not send value block, ignore
		//	- NBChanSendBlock is the only wait where thread has value
		if chanState != NBChanSendBlock {
			return // thread is not held in send value
		}
	}
	// thread holds value in state NBChanSendBlock

	// because this thread holds outputLock,
	// only one thread at a time may arrive here
	//	- competing with consumers and closeNow for the value

	// ensure two-chan receive operation is available
	if !n.tcSendBlock.Load() {
		return // the value went to another thread
	}

	// prepare two-chan receive second channel
	var collectChan = n.collectChan.Get(1)
	if len(collectChan) > 0 {
		<-collectChan
	}
	n.collectChanActive.Store(&collectChan)

	// seek permission for two-chan receive
	if !n.tcSendBlock.CompareAndSwap(true, false) {
		return // the value went to another thread
	}

	// two-chan fetch of value
	select {
	case value, hasValue = <-n.closableChan.Ch():
	case <-collectChan:
	}

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
