/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// WinOrWaiter picks a winner thread to carry out some task used by many threads.
//   - threads arriving to an idle WinorWaiter are winners that complete the task
//   - After a winning thread completes the task, it invokes WinnerDone
//   - threads arriving to a WinOrWait in progress are held waiting until WinnerDone
//   - the task is completed on demand, but only by the first thread requesting it
type WinOrWaiter struct {
	WinOrWaiterCore
}

// WinOrWaiter returns a semaphore used for completing an on-demand task by
// the first thread requesting it, and that result shared by subsequent threads held
// waiting for the result.
func NewWinOrWaiter(mustBeLater bool, g0 GoGen) (winOrWaiter *WinOrWaiter) {
	w := WinOrWaiter{WinOrWaiterCore: *NewWinOrWaiterCore(mustBeLater, g0.Go())}
	go w.contextThread()
	return &w
}

func (ww *WinOrWaiter) contextThread() {
	g0 := ww.WinOrWaiterCore.g0
	var err error
	defer g0.Done(&err)
	defer Recover(Annotation(), &err, NoOnError)

	<-g0.Context().Done()
	ww.Broadcast()
}
