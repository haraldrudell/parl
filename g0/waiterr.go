/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
)

// waiterr is a private combined error channel and waitgroup.
// parl.NBChan is a non-blocking channel.
// waiter is an observable WaitGroup such as parl.WaitGroup or
// parl.TraceGroup
// public is exactly Ch() IsExit() Wait() String().
// waiterr is initialized in a composite literal:
//  waiterr{wg: &parl.WaitGroup{}}
type waiterr struct {
	errCh parl.NBChan[parl.GoError]
	wg    waiter
}

func (we *waiterr) Ch() (ch <-chan parl.GoError) {
	return we.errCh.Ch()
}

func (we *waiterr) IsExit() (isExit bool) {
	return we.wg.IsZero()
}

func (we *waiterr) Wait() {
	we.wg.Wait()
}

func (we *waiterr) String() (s string) {
	return we.wg.String()
}

func (we *waiterr) send(goError parl.GoError) {
	we.errCh.Send(goError)
}

func (we *waiterr) close() {
	we.errCh.Close()
}

func (we *waiterr) add(delta int) {
	we.wg.Add(delta)
}

func (we *waiterr) done() {
	we.wg.Done()
}

func (we *waiterr) doneBool() (isExit bool) {
	return we.wg.DoneBool()
}

func (we *waiterr) counters() (adds, dones int) {
	return we.wg.Counters()
}

func (we *waiterr) didClose() (diClose bool) {
	return we.errCh.DidClose()
}
