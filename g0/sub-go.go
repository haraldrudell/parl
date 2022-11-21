/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	subGoFrames = 1
)

type SubGo struct {
	g0               parl.Go // Register()
	local            bool
	waiter           // Wait() String()
	isNonFatal       parl.AtomicBool
	isErrorExit      parl.AtomicBool
	cancelAndContext // Cancel() Context()
}

func NewGoSub(g0 parl.Go, local ...parl.GoSubLocal) (subGo parl.SubGo) {
	var local0 bool
	if len(local) > 0 {
		local0 = local[0] == parl.GoSubIsLocal
	}
	return &SubGo{
		g0:               g0,
		local:            local0,
		waiter:           &parl.TraceGroup{},
		cancelAndContext: *newCancelAndContext(g0.Context()),
	}
}

func (gc *SubGo) Register() {
	gc.g0.Register()
}

func (gc *SubGo) Add(delta int) {
	gc.waiter.Add(delta)
	if !gc.local {
		gc.g0.Add(delta)
	}
}

func (gc *SubGo) AddError(err error) {
	if err != nil {
		if !perrors.HasStack(err) {
			err = perrors.Stackn(err, subGoFrames)
		}
		gc.isNonFatal.Set()
	}
	gc.g0.AddError(err)
}

func (gc *SubGo) Done(errp *error) {
	if errp != nil {
		if *errp != nil {
			gc.isErrorExit.Set()
		}
		if !perrors.HasStack(*errp) {
			*errp = perrors.Stackn(*errp, subGoFrames)
		}
	}
	gc.waiter.Done() // done without sending error
	if gc.local {
		if errp != nil && *errp != nil {
			gc.AddError(*errp)
		}
	} else {
		gc.g0.Done(errp)
	}
}

func (gc *SubGo) IsErr() (isNonFatal bool, isErrorExit bool) {
	return gc.isNonFatal.IsTrue(), gc.isErrorExit.IsTrue()
}

func (gc *SubGo) SubGo(local ...parl.GoSubLocal) (goCancel parl.SubGo) {
	return NewGoSub(gc, local...)
}
