/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
)

// Send sends a single value on the channel
//   - non-blocking, thread-safe, panic-free and error-free
func (n *NBChan[T]) Send(value T) {
	if n.isCloseInvoked.Load() {
		return // no send after Close(), atomic performance: noop
	}
	n.preSend()
	n.inputLock.Lock()
	defer n.postSend()

	if n.isCloseInvoked.Load() {
		return // no send after Close() return: noop
	}

	// try to provide value to thread
	if !n.noThread.Load() && n.alertOrLaunchThreadWithValue(value) {
		return // value was handed to thread return: done
	}

	// save value in [NBChan.inputQueue]
	if n.inputQueue == nil {
		n.inputQueue = n.newQueue(1) // enforce allocation size
	}
	n.inputQueue = append(n.inputQueue, value)
	atomic.StoreUint64(&n.inputCapacity, uint64(cap(n.inputQueue)))
	atomic.AddUint64(&n.unsentCount, 1)
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

	if !n.noThread.Load() && n.alertOrLaunchThreadWithValue(values[0]) {
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
	atomic.StoreUint64(&n.inputCapacity, uint64(cap(n.inputQueue)))
	atomic.AddUint64(&n.unsentCount, uint64(valueCount))
}

// preSend registers a Send/SendMany invocation pre-inputLock
func (n *NBChan[T]) preSend() {
	if atomic.AddUint64(&n.sends, 1) == 1 {
		n.sendsWait.HoldWaiters()
	}
}

// post send is deferred for [NBChan.Send] and [NBChan.SendMany]
//   - release inputLock
//   - alert thread if no pending Get and values ar present
func (n *NBChan[T]) postSend() {
	n.inputLock.Unlock()

	// decrement sends
	var sends = int(atomic.AddUint64(&n.sends, ^uint64(0)))

	// update data available
	var unsentCount = atomic.LoadUint64(&n.unsentCount)
	n.setDataAvailable(unsentCount > 0)

	if n.noThread.Load() {
		return // no thread progress to manage
	}

	// ensure thread progress

	// if no data only release any sends waiters
	if unsentCount == 0 {
		// release thread if it is waiting for pending sends
		if sends == 0 {
			n.sendsWait.ReleaseWaiters()
		}
		return
	}

	// a send ended and data is available
	// release any sends waiters to get the new data
	n.sendsWait.ReleaseWaiters()
	// alert any waiting always-threads to get the new data
	n.alertThread(nil)
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
	var size = int(atomic.LoadUint64(&n.allocationSize))
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

// alertOrLaunchThreadWithValue attempts to launch thread using value
//   - invoked while holding [NBChan.inputLock]
//   - Invoked by Send/SendMany
//   - value has nit been added to unsentCount yet
func (n *NBChan[T]) alertOrLaunchThreadWithValue(value T) (didProvideValue bool) {

	// if Get in progress or NBChan is not empty, don’t launch thread
	if atomic.LoadUint64(&n.gets) > 0 || // [NBChan.Get] in progress
		atomic.LoadUint64(&n.unsentCount) > 0 { // channel is not empty
		return // gets are in progress or not empty return: no thread interaction allowed
	}

	if n.alertThread(&value) {
		didProvideValue = true
		return
	} else if didProvideValue = n.tryStartThread(); !didProvideValue {
		return // this invocation did not win thread-launch return : didProvideValue: false
	}

	atomic.AddUint64(&n.unsentCount, 1)
	n.startThread(value)
	return // launched new thread return: didProvideValue: true
}

// tryStartThread attempts to be the thread that gets to launch sendThread
//   - inside threadLock
//   - on isRunningThread true, closesOnThreadSend must be valid
func (n *NBChan[T]) tryStartThread() (isWinner bool) {
	var ch chan struct{}
	var didWin bool
	if n.closesOnThreadSend.Load() == nil {
		ch = make(chan struct{})
		didWin = n.closesOnThreadSend.CompareAndSwap(nil, &ch)
	}
	n.threadLock.Lock()
	defer n.threadLock.Unlock()

	// check for thread already running or this invocation should not launch it
	if isWinner = n.isRunningThread.CompareAndSwap(false, true); !isWinner {
		return // thread was already running return
	}
	if !didWin {
		if ch == nil {
			ch = make(chan struct{})
		}
		n.closesOnThreadSend.Store(&ch) // fresh channel
	}
	return
}

// startThread launches the send thread
//   - on Send/SendMany when unsentCount was 0 and no gets in progress
//   - on postGet if unsentCount > 0 and gets in progress went to 0
//   - isRunningThread.CompareAndSwap winner invokes startThread
//   - thread runs until unsentCount is 0 inside threadLock
func (n *NBChan[T]) startThread(value T) {
	var endCh = make(chan struct{})
	n.threadWait.Store(&endCh)
	go n.sendThread(value) // send err in new thread
}

func (n *NBChan[T]) waitForSendThread() {
	var ch chan struct{}
	if chp := n.threadWait.Load(); chp == nil {
		return
	} else {
		ch = *chp
	}

	<-ch
}

// alertThread alerts any waiting always-threads
//   - invoked from Send/SendMany
//   - value has not been added to unsentCount yet
func (n *NBChan[T]) alertThread(valuep *T) (didAlert bool) {
	if didAlert = n.threadAlertPending.CompareAndSwap(true, false); !didAlert {
		return // sendThread not waiting for alert or this invocation not winner return
	}

	// alert sendThread
	if valuep != nil {
		atomic.AddUint64(&n.unsentCount, 1)
		n.threadAlertValue.Store(valuep)
	}
	n.threadAlertWait.ReleaseWaiters()

	return
}

func (n *NBChan[T]) updateDataAvailable() (dataCh chan struct{}) {
	if n.closableChan.IsClosed() {
		return n.setDataAvailableAfterClose()
	}
	return n.setDataAvailable(atomic.LoadUint64(&n.unsentCount) > 0)
}

func (n *NBChan[T]) setDataAvailableAfterClose() (dataCh chan struct{}) {
	return n.setDataAvailable(true)
}

func (n *NBChan[T]) setDataAvailable(isAvailable bool) (dataCh chan struct{}) {
	if chp := n.dataWaitCh.Load(); chp != nil && n.isDataAvailable.Load() == isAvailable {
		dataCh = *chp
		return // initialized and in correct state return: noop
	}
	n.availableLock.Lock()
	defer n.availableLock.Unlock()

	// not yet initialized case
	var chp = n.dataWaitCh.Load()
	if chp == nil {
		dataCh = make(chan struct{})
		if isAvailable {
			close(dataCh)
		}
		n.dataWaitCh.Store(&dataCh)
		n.isDataAvailable.Store(isAvailable)
		return // channel initialized and state set return
	}

	// is state correct?
	dataCh = *chp
	if n.isDataAvailable.Load() == isAvailable {
		return // channel was initialized and state was correct return
	}

	// should channel be closed: if data is available
	if isAvailable {
		close(*chp)
		n.isDataAvailable.Store(true)
		return // channel closed andn state updated return
	}

	// replace with open channel: data is not available
	dataCh = make(chan struct{})
	n.dataWaitCh.Store(&dataCh)
	n.isDataAvailable.Store(false)
	return // new open channel stored and state updated return
}
