/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
)

type SubGo struct {
	parl.Go          // Register() AddError()
	waiter           // Wait() String()
	cancelAndContext // Cancel() Context()
}

func NewGoSub(g0 parl.Go) (subGo parl.SubGo) {
	return &SubGo{
		Go:               g0,
		waiter:           &parl.TraceGroup{},
		cancelAndContext: *newCancelAndContext(g0.Context()),
	}
}

func (gc *SubGo) Add(delta int) {
	gc.waiter.Add(delta)
	gc.Go.Add(delta)
}

func (gc *SubGo) Done(errp *error) {
	gc.waiter.Done() // done without sending error
	gc.Go.Done(errp)
}

func (gc *SubGo) SubGo() (goCancel parl.SubGo) {
	return NewGoSub(gc)
}
