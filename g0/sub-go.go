/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"sync"

	"github.com/haraldrudell/parl"
)

type GoSub struct {
	parl.Go
	wg  sync.WaitGroup
	ctx parl.CancelContext
}

func NewGoSub(g0 parl.Go) (goCancel parl.SubGo) {
	return &GoSub{Go: g0, ctx: parl.NewCancelContext(g0.Context())}
}

func (gc *GoSub) Add(delta int) {
	gc.wg.Add(delta)
	gc.Go.Add(delta)
}

func (gc *GoSub) Done(err error) {
	gc.wg.Done()
	gc.Go.Done(err)
}

func (gc *GoSub) Wait() {
	gc.wg.Wait()
}

func (gc *GoSub) Cancel() {
	gc.ctx.Cancel()
}

func (gc *GoSub) SubGo() (goCancel parl.SubGo) {
	return NewGoSub(gc)
}
