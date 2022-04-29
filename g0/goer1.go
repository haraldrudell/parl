/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/psync"
)

type Goer1 struct {
	errCh parl.NBChan[parl.GoError]
	wg    psync.TraceGroup
	ctx   parl.CancelContext
}

func NewGoer1(ctx context.Context) (goer parl.SubGoer) {
	return &Goer1{ctx: parl.NewCancelContext(ctx)}
}

func (gr *Goer1) Go() (g0 parl.Go) {
	gr.wg.Add(1)
	return NewGo(
		gr.errorReceiver,
		gr.wg.Add,
		gr.done,
		gr.ctxFn,
	)
}

func (gr *Goer1) Ch() (ch <-chan parl.GoError) {
	return gr.errCh.Ch()
}

func (gr *Goer1) AddError(err error) {
	gr.errCh.Send(NewGoError(
		err,
		parl.GeNonFatal,
		gr,
	))
}

func (gr *Goer1) Context() (ctx context.Context) {
	return gr.ctx
}

func (gr *Goer1) Cancel() {
	gr.ctx.Cancel()
}

func (gr *Goer1) Wait() {
	gr.wg.Wait()
}

func (gr *Goer1) IsExit() (isExit bool) {
	return gr.wg.IsZero()
}

func (gr *Goer1) String() (s string) {
	return gr.wg.String()
}

func (gr *Goer1) errorReceiver(err error) {
	if err == nil {
		return
	}
	gr.errCh.Send(NewGoError(err, parl.GeNonFatal, gr))
}

func (gr *Goer1) ctxFn() (ctx context.Context) {
	return gr.ctx
}

func (gr *Goer1) done(err error) {
	isZero := gr.wg.DoneBool()

	var source parl.GoErrorSource
	if isZero {
		source = parl.GeExit
	} else {
		source = parl.GePreDoneExit
	}
	gr.errCh.Send(NewGoError(err, source, gr))

	if isZero {
		gr.errCh.Close()
	}
}
