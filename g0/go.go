/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package g0 facilitates launch, management and execution of goroutines
package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

type GoDo struct {
	addError func(err error)
	add      func(delta int)
	done     func(err *error)
	context  func() (ctx context.Context)
	cancel   func()
}

func NewGo(
	errorReceiver func(err error),
	add func(delta int),
	done func(err *error),
	context func() (ctx context.Context),
	cancel func()) (g0 parl.Go) {
	return &GoDo{
		addError: errorReceiver,
		add:      add,
		done:     done,
		context:  context,
		cancel:   cancel,
	}
}

func (g0 *GoDo) Register() {
	g0.add(0)
}

func (g0 *GoDo) Add(delta int) {
	g0.add(delta)
}

func (g0 *GoDo) AddError(err error) {
	g0.addError(err)
}

func (g0 *GoDo) Done(errp *error) {
	g0.done(errp)
}

func (g0 *GoDo) Context() (ctx context.Context) {
	return g0.context()
}

func (g0 *GoDo) Cancel() {
	g0.cancel()
}

func (g0 *GoDo) SubGo() (goCancel parl.SubGo) {
	return NewGoSub(g0)
}
