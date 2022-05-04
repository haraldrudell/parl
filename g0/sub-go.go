/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
)

type GoSub struct {
	parl.Go
	waiter           // Wait() String()
	cancelAndContext // Cancel() Context()
}

func NewGoSub(g0 parl.Go) (goCancel parl.SubGo) {
	return &GoSub{
		Go:               g0,
		cancelAndContext: *newCancelAndContext(g0.Context()),
	}
}

func (gc *GoSub) Add(delta int) {
	gc.waiter.Add(delta)
	gc.Go.Add(delta)
}

func (gc *GoSub) Done(errp *error) {
	gc.waiter.Done() // done without sending error
	gc.Go.Done(errp)
}

func (gc *GoSub) SubGo() (goCancel parl.SubGo) {
	return NewGoSub(gc)
}
