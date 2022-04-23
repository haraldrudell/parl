/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

type GoDo struct {
	errorReceiver func(err error)
	add           func(delta int)
	done          func(err error)
	ctx           context.Context
}

func NewGo(
	errorReceiver func(err error),
	add func(delta int),
	done func(err error),
	ctx context.Context) (g0 parl.Go) {
	return &GoDo{
		errorReceiver: errorReceiver,
		add:           add,
		done:          done,
		ctx:           ctx,
	}
}

func (g0 *GoDo) Register() {
	g0.add(0)
}

func (g0 *GoDo) Add(delta int) {
	g0.add(delta)
}

func (g0 *GoDo) AddError(err error) {
	g0.errorReceiver(err)
}

func (g0 *GoDo) Done(err error) {
	g0.done(err)
}

func (g0 *GoDo) Context() (ctx context.Context) {
	return g0.ctx
}
