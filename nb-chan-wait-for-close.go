/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// WaitForClose blocks until the channel is closed and empty
//   - if Close is not invoked or the channel is not read to end,
//     WaitForClose blocks indefinitely
//   - if CloseNow is invoked, WaitForClose is unblocked
//   - if errp is non-nil, any thread and close errors are appended to it
//   - a close error will already have been returned by Close
//   - thread-safe, panic-free, deferrable
func (n *NBChan[T]) WaitForClose(errp ...*error) {
	defer n.appendErrors(nil, errp...)

	<-n.doWaitForClose()
}

// WaitForCloseCh returns a channel that closes when [NBChan.Ch] closes
//   - the underlying channel is sending items, this channel is not
func (n *NBChan[T]) WaitForCloseCh() (ch <-chan struct{}) {
	return n.doWaitForClose()
}

// initWaitForClose ensures that waitForClose is valid
func (n *NBChan[T]) doWaitForClose() (ch <-chan struct{}) {

	// ensure channel present
	var ch0 chan struct{}
	if chp := n.waitForClose.Load(); chp != nil {
		ch0 = *chp
	} else {
		ch0 = make(chan struct{})
		if !n.waitForClose.CompareAndSwap(nil, &ch0) {
			ch0 = *n.waitForClose.Load()
		}
	}
	ch = ch0

	// close ch if nb.ClosableChan.Ch() is closed
	if !n.closableChan.IsClosed() ||
		!n.isWaitForCloseDone.CompareAndSwap(false, true) {
		return // not close yet or not winner return
	}

	close(ch0) // once executed Done

	return
}
