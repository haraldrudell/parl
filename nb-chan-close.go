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

	// atomic performance check if Close already invoked
	if n.isCloseInvoked.IsInvoked() {
		// await winner return
		<-n.isCloseInvoked.Ch()
		return // Close complete return
	}

	// select Close winner
	var isWinner, isRunningThread, done = n.selectCloseWinner()
	if !isWinner {
		// await winner return
		<-n.isCloseInvoked.Ch()
		return
	}
	defer done.Done()
	// isCloseInvoked.IsInvoked is true so new Send SendMany will return immediately
	//	- unsent count is therefore strictly decreasing by Get and send-thread

	// handle deferred Close
	if isRunningThread {
		// a thread is running, if it does not exit, it’s deferred close
		//	- await static thread state, one of: send-block exit sends gets alert
		var threadState NBChanTState
		select {
		// thread exit
		case <-n.tcThreadExitAwaitable.Ch():
			// the thread exited
			// if at end of items, executee close immediately
			if n.unsentCount.Load() == 0 {
				threadState = NBChanExit
			}

			// NBChanSendBlock NBChanAlert NBChanSends NBChanGets
		case threadState = <-n.stateCh():
		}

		// an always thread in NBChanAlert must be alerted
		//	- isCloseInvoked.IsInvoked prevents further alert wait
		if threadState == NBChanAlert {
			n.tcAlertThread()
		}

		// deferred close function
		if threadState != NBChanExit {
			return // deferred close
		}
	}

	// immediate close
	//	- send-thread or Get to consume remaining items
	for n.unsentCount.Load() > 0 {
		<-n.updateDataAvailable()
		n.getsWait.Wait()
	}

	didClose, _ = n.executeChClose()
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
	var isWinner, isRunningThread, done, doneClose = n.selectCloseNowWinner()
	if !isWinner {
		<-n.isCloseNow.Ch()
		return // close now loser threads
	}
	// this is CloseNow winner thread
	defer done.Done()
	if doneClose != nil {
		defer doneClose.Done()
	}

	// wait for any Send SendMany Get to complete
	//	- Get Collect uses underlying channel
	//	- new invocations are canceled by:
	//	- — Send SendMany: isCloseInvoked.IsInvoked
	//	- — Get: isCloseNow.IsInvoked
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
			n.tcAlertThread()
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

	// close data ch waiter
	//	- this will release a possible held Close invocation
	n.setDataAvailableAfterClose()

	// execute close
	didClose, _ = n.executeChClose()

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
	done, doneClose Done,
) {
	// atomize closeNow winner selection with:
	//	- retrieving running thread state and
	//	- setting isCloseInvoked true
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	// select CloseNow winner
	if isWinner, done = n.isCloseNow.IsWinner(NoOnceWait); !isWinner {
		return // CloseNow was completed by another thread
	}
	// is winning CloseNow thread

	// CloseNow also signals Close
	_, doneClose = n.isCloseInvoked.IsWinner(NoOnceWait)

	// thread status at time of CloseNow
	isRunningThread = n.tcRunningThread.Load()

	return
}

// selectCloseWinner selects the winner to execuite close possily deferred to thread
//   - executeCloseNow true: is winner thread and close is not deferred
//   - deferred close is setting isCloseInvoked to true while a thread is running
//   - caller must hold inputLock for isCloseInvoked update
func (n *NBChan[T]) selectCloseWinner() (isWinner, isRunningThread bool, done Done) {
	// atomize close win with reading running thread-state
	n.tcThreadLock.Lock()
	defer n.tcThreadLock.Unlock()

	var _ OnceCh

	// select winner
	//	- losers do not wait here to get out of tcThreadLock
	if isWinner, done = n.isCloseInvoked.IsWinner(NoOnceWait); !isWinner {
		return // Close was already invoked return: executeCloseNow: false
	}
	// this thread is close winner

	isRunningThread = n.tcRunningThread.Load()

	return // if thread is not running: executeCloseNow: true: execute close now
}

// executeChClose closes the underlying channel
//   - didClose true: this invocation closed the channel
//   - err possible error, already submitted: unused
//   - idempotent thread-safe
//   - isCloseInvoked.IsInvoked must be true
//   - unsent count must be zero
//   - invoked by:
//   - — CloseNow
//   - — Close if not deferred close
//   - — send thread in deferred close: on exit if Close was invoked prior to thread exit
func (n *NBChan[T]) executeChClose() (didClose bool, err error) {

	// await Send SendMany ceasing
	//	- new invocations return immediately due to isCloseInvoked.IsInvoked true
	n.sendsWait.Wait()

	// wait for any Send SendMany Get to complete
	//	- Get Collect uses underlying channel

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
