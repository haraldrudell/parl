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

// tcEnsureThreadProgress ensures that thread is not exiting
func (n *NBChan[T]) tcEnsureThreadProgress() {
	n.threadProgressLock.Lock()
	defer n.threadProgressLock.Unlock()
	defer n.threadProgressRequired.Store(false)

	//	- unsent count was at 0, then incremented
	//	- on-demand or always thread
	//	- thread was launched at one time
	//	- all Send SendMany Get completed

	// await thread static state
	select {
	// thread exit
	case <-n.tcThreadExitAwaitable.Ch():
		if n.isThreadAlways.Load() {
			return // exiting always thread: noop
		}
		// NBChanSendBlock NBChanAlert NBChanGets NBChanSends
	case threadState := <-n.stateCh():
		if threadState == NBChanAlert {
			// unblock always thread awaiting alert
			n.tcAlertThread(nil)
		}
		return // thread holding somewhere return: ok
	}
	// thread did exit, it is on-demand thread

	// unless unsent count is again zero, the thread must not exit

	// close now or out of items
	if n.unsentCount.Load() == 0 || n.isCloseNow.IsInvoked() {
		return // out of items return: thread exit ok
	}

	// check close status
	if n.isCloseComplete() {
		return // underlying channel has closed: noop
	}

	// seek permisssion to start thread
	if !n.tcStartThreadWinner() {
		return // already restarted return
	}
	// start thread without value
	var value T
	var hasValue bool
	n.tcStartThread(value, hasValue)
}

func (n *NBChan[T]) isCloseComplete() (isClosed bool) {
	n.inputLock.Lock()
	defer n.inputLock.Unlock()

	isClosed = n.closableChan.IsClosed()

	return
}

// tcAlertThread alerts any waiting always-threads
//   - invoked from Send/SendMany
//   - value has not been added to unsentCount yet
func (n *NBChan[T]) tcAlertThread(valuep *T) (didAlert bool) {

	// is aleter armed?
	if !n.threadAlertWait.IsHold() {
		return // no
	}

	// alert sendThread
	if valuep != nil {
		n.unsentCount.Add(1)
		n.threadAlertValue.Store(valuep)
	}
	n.threadAlertWait.ReleaseWaiters()
	return
}

func (n *NBChan[T]) tcDeferredProgressCheck() (isDeferred bool) {
	n.tcProgressCheck.Lock()
	defer n.tcProgressCheck.Unlock()

	if isDeferred = n.gets.Load() > 0; !isDeferred {
		return
	}
	n.threadProgressRequired.Store(true)
	return
}

func (n *NBChan[T]) tcIsDeferredProgress() (needProgress bool) {
	n.tcProgressCheck.Lock()
	defer n.tcProgressCheck.Unlock()

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

// Thread status values
const (
	NBChanExit      NBChanTState = "exit"     // NBChan thread is not running
	NBChanAlert     NBChanTState = "alert"    // NBChan thread is always running and blocked idle waiting for alert
	NBChanGets      NBChanTState = "GetWait"  // NBChan thread is blocked waiting for Get invocations to complete
	NBChanSends     NBChanTState = "SendWait" // NBChan thread is blocked waiting for Send/SendMany invocations to complete
	NBChanSendBlock NBChanTState = "chSend"   // NBChan thread is blocked in channel send
	NBChanRunning   NBChanTState = "run"      // NBChan is running
	NBChanNoLaunch  NBChanTState = "none"     // thread was never launched
)

// state of NBChan thread
//   - NBChanExit NBChanAlert NBChanGets NBChanSends NBChanSendBlock
//     NBChanRunning
type NBChanTState string
