/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"fmt"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// waiterr is a private combined error channel and waitgroup.
// parl.NBChan is a non-blocking channel.
// waiter is an observable WaitGroup such as parl.WaitGroup or
// parl.TraceGroup.
// public is exactly Ch() IsExit() Wait() String().
// waiterr is initialized in a composite literal:
//  waiterr{wg: &parl.WaitGroup{}}
type waiterr struct {
	errCh parl.NBChan[parl.GoError]
	wg    waiter
	index parl.GoIndex
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
	adds, dones := we.wg.Counters()
	return fmt.Sprintf("%d(%d)", adds-dones, adds)
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

func (we *waiterr) string(errp ...*error) (s string) {
	if len(errp) > 0 {
		if errp0 := errp[0]; errp0 != nil && *errp0 != nil {
			s = " err: " + perrors.Short(*errp0)
		}
	}
	adds, dones := we.counters()
	return fmt.Sprintf("#%d %d(%d)%s", we.index, adds-dones, adds, s)
}
