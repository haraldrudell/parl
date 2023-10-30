/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
)

const (
// dataWaiterAfterClose = true
)

// Close orders the channel to close once pending sends complete.
//   - Close is thread-safe, non-blocking, error-free and panic-free
//   - when Close returns, the channel may still be open and have items
//   - didClose is true if this Close invocation actually did close the channel
//   - didClose may be false for all invocations if the channel is closed by sendThread
func (n *NBChan[T]) Close() (didClose bool) {
	if n.isCloseInvoked.Load() {
		return // Close was already invoked atomic performance
	}
	n.inputLock.Lock()
	defer n.inputLock.Unlock()

	// determine if this invocation is the closing
	//	- secondary, if thread is running detect deferred close
	if !n.tryClose() {
		return // not this invocation or close deferred to running thread return
	}

	// do the close
	didClose, _ = n.close() // invocation is holding inputLock

	// close datawaiter
	n.setDataAvailableAfterClose()
	return
}

// CloseNow closes without waiting for sends to complete.
//   - CloseNow is thread-safe, panic-free, idempotent, deferrable and is
//     designed to not block for long
//   - CloseNow does not return until the channel is closed and no thread is running
//   - Upon return, errp and err receive any close or panic errors for this [NBChan]
//   - if errp is non-nil, it is updated with error status
func (n *NBChan[T]) CloseNow(errp ...*error) (didClose bool, err error) {
	n.outputLock.Lock()
	defer n.outputLock.Unlock()
	n.inputLock.Lock()
	defer n.inputLock.Unlock()
	defer n.appendErrors(&err, errp...) // add any errors for this [NBChan]

	// check for closeNow already completed
	var doCloseNow, isRunningThread bool
	if doCloseNow, isRunningThread = n.tryCloseNow(); !doCloseNow {
		return // CloseNow already completed return: noop
	}

	// release thread so it will exit
	//	- if sendThread is blocked in channel send
	//	- while another thread closes the channel
	//	- that is a data race
	if isRunningThread {
		// empty the channel if thread is in channel send
		n.discardSendThreadValue()
		// alert any holding always-threads
		n.alertThread(nil)
	}

	// execute close
	didClose, _ = n.close() // invocation is holding inputLock

	// discard pending data
	if nT := len(n.inputQueue); nT > 0 {
		atomic.AddUint64(&n.unsentCount, uint64(-nT))
	}
	n.inputQueue = nil
	atomic.StoreUint64(&n.inputCapacity, 0)
	if nT := len(n.outputQueue); nT > 0 {
		atomic.AddUint64(&n.unsentCount, uint64(-nT))
	}
	n.outputQueue = nil
	atomic.StoreUint64(&n.outputCapacity, 0)

	// await thread exit
	n.waitForSendThread()

	// close data ch waiter
	n.setDataAvailableAfterClose()

	return
}

// tryCloseNow tries to win didClose execution
//   - CloseNow holdsinputLock for isClose changes
//   - CloseNow holds outputLock and inputLock so now thread launch after this
func (n *NBChan[T]) tryCloseNow() (
	doCloseNow,
	isRunningThread bool,
) {
	n.threadLock.Lock()
	defer n.threadLock.Unlock()

	n.isCloseInvoked.CompareAndSwap(false, true)                                   // invocation is holding inputLock
	if doCloseNow = n.isCloseNowInvoked.CompareAndSwap(false, true); !doCloseNow { // invocation is holding inputLock
		return
	}
	isRunningThread = n.isRunningThread.Load()

	return
}

// tryClose attempts to win isCloseInvoked and detect if close invocation should be deferred
//   - Close holds inputLock for isClose changes
func (n *NBChan[T]) tryClose() (executeCloseNow bool) {
	n.threadLock.Lock()
	defer n.threadLock.Unlock()

	// invocation is holding inputLock
	if !n.isCloseInvoked.CompareAndSwap(false, true) {
		return // Close was already invoked return: executeCloseNow: false
	}
	if executeCloseNow = !n.isRunningThread.Load(); !executeCloseNow {
		// alert any waiting always-threads
		n.alertThread(nil)
	}

	return // if thread is not running: executeCloseNow: true: execute close now
}

// close closes the underlying channel
//   - invoked when isCloseInvoked true.
//   - isCloseInvoked	was set while holding inputLock
func (n *NBChan[T]) close() (didClose bool, err error) {
	if didClose, err = n.closableChan.Close(); !didClose {
		return // already closed return: noop
	} else if err != nil {
		n.AddError(err) // store possible close error
	}
	n.doWaitForClose() // update [NBChan.waitForClose]

	return
}

// appendErrors aggregates any errors for this [NBChan] in any non-nil errp0 or errp
//   - like perrors.AppendErrorDefer but allows for errp to be nil
func (n *NBChan[T]) appendErrors(errp0 *error, errp ...*error) {

	if errp0 != nil {
		perrors.AppendErrorDefer(errp0, nil, n.GetError)
	}
	// obtain error pointers
	for _, errpx := range errp {
		if errpx == nil {
			continue
		}
		perrors.AppendErrorDefer(errpx, nil, n.GetError)
	}
}

// discardThreadValue ends any sendThread channel send
func (n *NBChan[T]) discardSendThreadValue() {
	var chp = n.closesOnThreadSend.Load()
	if chp == nil {
		return // thread not initialized return: hasValue false
	}
	select {
	case <-*chp:
	case <-n.closableChan.Ch():
	}
}
