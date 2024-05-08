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

	<-n.waitForClose.Ch()
}

// WaitForCloseCh returns a channel that closes when [NBChan.Ch] closes
//   - the underlying channel is sending items, this channel is not
func (n *NBChan[T]) WaitForCloseCh() (ch AwaitableCh) { return n.waitForClose.Ch() }
