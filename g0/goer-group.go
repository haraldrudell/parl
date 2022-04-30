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

type GoerGroup struct {
	errCh parl.NBChan[parl.GoError]
	wg    psync.TraceGroup
	ctx   parl.CancelContext
}

func NewGoerGroup(ctx context.Context) (goer parl.GoerGroup) {
	return &GoerGroup{ctx: parl.NewCancelContext(ctx)}
}

func (gr *GoerGroup) Go() (g0 parl.Go) {
	gr.wg.Add(1)
	return NewGo(
		gr.errorReceiver,
		gr.wg.Add,
		gr.done,
		gr.ctxFn,
	)
}

func (gr *GoerGroup) Ch() (ch <-chan parl.GoError) {
	return gr.errCh.Ch()
}

func (gr *GoerGroup) AddError(err error) {
	gr.errCh.Send(NewGoError(
		err,
		parl.GeNonFatal,
		gr,
	))
}

func (gr *GoerGroup) Context() (ctx context.Context) {
	return gr.ctx
}

func (gr *GoerGroup) Cancel() {
	gr.ctx.Cancel()
}

func (gr *GoerGroup) Wait() {
	gr.wg.Wait()
}

func (gr *GoerGroup) IsExit() (isExit bool) {
	return gr.wg.IsZero()
}

func (gr *GoerGroup) String() (s string) {
	return gr.wg.String()
}

func (gr *GoerGroup) errorReceiver(err error) {
	if err == nil {
		return
	}
	gr.errCh.Send(NewGoError(err, parl.GeNonFatal, gr))
}

func (gr *GoerGroup) ctxFn() (ctx context.Context) {
	return gr.ctx
}

func (gr *GoerGroup) done(err error) {
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
