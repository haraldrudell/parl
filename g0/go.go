/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package g0 facilitates launch, management and execution of goroutines
package g0

import (
	"context"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	goFrames = 1
)

type Go struct {
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
	return &Go{
		addError: errorReceiver,
		add:      add,
		done:     done,
		context:  context,
		cancel:   cancel,
	}
}

func (g0 *Go) Register() {
	g0.add(0)
}

func (g0 *Go) Add(delta int) {
	g0.add(delta)
}

func (g0 *Go) AddError(err error) {
	if err != nil && !perrors.HasStack(err) {
		err = perrors.Stackn(err, goFrames)
	}
	g0.addError(err)
}

func (g0 *Go) Done(errp *error) {
	if errp != nil && *errp != nil && !perrors.HasStack(*errp) {
		*errp = perrors.Stackn(*errp, goFrames)
	}
	g0.done(errp)
}

func (g0 *Go) Context() (ctx context.Context) {
	return g0.context()
}

func (g0 *Go) Cancel() {
	g0.cancel()
}

func (g0 *Go) SubGo(local ...parl.GoSubLocal) (goCancel parl.SubGo) {
	return NewGoSub(g0, local...)
}
