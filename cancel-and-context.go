/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "context"

const (
	ccSkipFrames = 1
	ccPrepend    = "— "
)

type CancelAndContext struct {
	ctx CancelContextDo
}

func NewCancelAndContext(ctx context.Context) (cc *CancelAndContext) {
	c := NewCancelContext(ctx)
	return &CancelAndContext{ctx: *c.(*CancelContextDo)}
}

func (cc *CancelAndContext) Cancel() {
	if IsThisDebugN(ccSkipFrames) {
		GetDebug(ccSkipFrames)("CancelAndContext.Cancel:\n" + newStack(ccSkipFrames).Shorts(ccPrepend))
	}
	cc.ctx.Cancel()
}

func (cc *CancelAndContext) Context() (ctx context.Context) {
	return cc.ctx.Context
}
