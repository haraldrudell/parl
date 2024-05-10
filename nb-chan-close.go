/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
)

// Close closes the underlying channel without data loss
//   - didClose true: this Close invocation executed channel close
//   - didClose may be false for all invocations if the channel is closed by sendThread
//   - when Close returns, the channel may still be open and have items
//   - Close is thread-safe, non-blocking, error-free and panic-free
//   - underlying channel closes once Send SendMany completes and the channel
//     is empty
func (n *NBChan[T]) Close() (didClose bool) {
	if n.isCloseInvoked.Load() {
		return // Close was already invoked atomic performance
	}
	// n.inputLock.Lock()
	// defer n.inputLock.Unlock()

	// determine if this invocation is the closing
	//	- secondary, if thread is running detect deferred close
	var isWinner, isRunningThread = n.selectCloseWinner()
	if !isWinner {
		return
	} else if isRunningThread {
		// a thread is running, so deferred close
		//	- await thread state like send block, exit or alert wait
		//	- holding inputLock and isCloseInvoked so no more items will be added
		//	- collect next thread state
		var threadState NBChanTState
		select {
		// thread exit
		case <-n.tcThreadExitAwaitable.Ch():
			// the thread exited

			// NBChanSendBlock NBChanAlert NBChanSends
		case threadState = <-n.stateCh():
		}

		// an always thread in NBChanAlert must be alerted
		if threadState == NBChanAlert {
			n.tcAlertNoValue()
		}

		return // not this invocation or close deferred to running thread return
	}

	// do the close
	didClose, _ = n.executeChClose() // invocation is holding inputLock
	// update datawaitCh
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

	// add any occuring errors for this [NBChan]
	defer n.appendErrors(&err, errp...)

	// select close now winner
	var isWinner, isRunningThread, done = n.selectCloseNowWinner()
	if !isWinner {
		return // close now loser threads
	}
	defer done.Done()

	// wait for any Send SendMany Get to complete
	//	- Get Collect uses underlying channel
	//	- new invocations are canceled by
	n.getsWait.Wait()
	n.sendsWait.Wait()

	// release thread so it will exit
	//	- gets and sends is zero, so thread is not held up there
	//	- isCloseNow.IsInvoked is true
	if isRunningThread {
		// a thread was running when IsCloseNow activated
		//	- await a state for the thread
		//	- holding inputLock outputLock so no Send SendMany Get
		var threadState NBChanTState
		select {
		// thread exit
		case <-n.tcThreadExitAwaitable.Ch():
			// NBChanSendBlock NBChanAlert
		case threadState = <-n.stateCh():
		}

		switch threadState {
		// alway thread awaiting alert
		case NBChanAlert:
			// alert and isclosenow will cause thread to exit
			n.tcAlertNoValue()
			// thread blocked in value send
		case NBChanSendBlock:
			select {
			// discard any data item received from thread
			//	- thread will exit due to isCloseNow
			case <-n.closableChan.Ch():
				// cancel on thread exit
			case <-n.tcThreadExitAwaitable.Ch():
			}
		}
	}

	// await thread exit
	if isRunningThread {
		<-n.tcThreadExitAwaitable.Ch()
	}

	// execute close
	//	- invocation is holding inputLock
	didClose, _ = n.executeChClose()

	// close data ch waiter
	n.setDataAvailableAfterClose()

	// discard pending data
	n.outputLock.Lock()
	defer n.outputLock.Unlock()
	n.inputLock.Lock()
	defer n.inputLock.Unlock()

	//	- thread has exited
	//	- Send SendMany Gets have ceased
	//	- this thread holds both locks

	if nT := len(n.inputQueue); nT > 0 {
		n.unsentCount.Add(uint64(-nT))
	}
	n.inputQueue = nil
	n.inputCapacity.Store(0)
	if nT := len(n.outputQueue); nT > 0 {
		n.unsentCount.Add(uint64(-nT))
	}
	n.outputQueue = nil
	n.outputCapacity.Store(0)

	return
}

// selectCloseNowWinner ensures CloseNow execution
//   - isWinner true: this thread executes close now
//   - isRunningThread: a thread was running at time of close now/
//     if so, tcThreadExitAwaitable is armed
//   - done: completion Done for winner thread
//   - caller must hold inputLock: to update isCloseNow and isCloseInvoked
//   - caller must hold inputLock and outputLock to prevent subsequent thread launch
func (n *NBChan[T]) selectCloseNowWinner() (
	isWinner,
	isRunningThread bool,
	done Done,
) {
	// atomize closeNow winner with running thread state
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	// select CloseNow winner
	if isWinner, done = n.isCloseNow.IsWinner(); !isWinner {
		return // CloseNow was completed by another thread
	} else {
		defer done.Done()
	}
	// is winning CloseNow thread

	// CloseNow also signals Close
	n.isCloseInvoked.CompareAndSwap(false, true)

	// thread stats at time of CloseNow
	isRunningThread = n.tcRunningThread.Load()
	return
}

// selectCloseWinner selects the winner to execuite close possily deferred to thread
//   - executeCloseNow true: is winner thread and close is not deferred
//   - deferred close is setting isCloseInvoked to true while a thread is running
//   - caller must hold inputLock for isCloseInvoked update
func (n *NBChan[T]) selectCloseWinner() (isWinner, isRunningThread bool) {
	// atomize close win with running thread-state
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	// select winner
	// invocation is holding inputLock
	if isWinner = n.isCloseInvoked.CompareAndSwap(false, true); !isWinner {
		return // Close was already invoked return: executeCloseNow: false
	}
	// this thread is close winner

	isRunningThread = n.tcRunningThread.Load()

	return // if thread is not running: executeCloseNow: true: execute close now
}

// executeChClose closes the underlying channel if not already closed
//   - didClose true if this invocation closed the channel
//   - err possible error, already submitted: unused
//   - TODO: invoker must hold inputLock or be sendThread
//   - invoked by:
//   - — CloseNow
//   - — Close if no thread is running
//   - — send thread on exit if Close was invoked prior to thread exit
func (n *NBChan[T]) executeChClose() (didClose bool, err error) {

	// wait for any Send SendMany Get to complete
	//	- Get Collect uses underlying channel
	n.getsWait.Wait()
	n.sendsWait.Wait()

	if didClose, err = n.closableChan.Close(); !didClose {
		return // already closed return: noop
	} else if err != nil {
		n.AddError(err) // store possible close error
	}
	// update [NBChan.waitForClose]
	n.waitForClose.Close()

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
