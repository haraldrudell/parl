/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "context"

type CancelAndContext struct {
	ctx CancelContextDo
}

func NewCancelAndContext(ctx context.Context) (cc *CancelAndContext) {
	c := NewCancelContext(ctx)
	return &CancelAndContext{ctx: *c.(*CancelContextDo)}
}

func (cc *CancelAndContext) Cancel() {
	cc.ctx.Cancel()
}

func (cc *CancelAndContext) Context() (ctx context.Context) {
	return cc.ctx.Context
}
