/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/goid"
)

const (
	ccSkipFrames = 1
	ccPrepend    = "— "
)

// cancelAndContext provides a private field that promotes Cancel and Context methods
type cancelAndContext struct {
	ctx context.Context
}

func newCancelAndContext(ctx context.Context) (cc *cancelAndContext) {
	return &cancelAndContext{ctx: parl.NewCancelContext(ctx)}
}

func (cc cancelAndContext) Cancel() {

	// if caller is debug, debug-print cancel action
	if parl.IsThisDebugN(ccSkipFrames) {
		parl.GetDebug(ccSkipFrames)("CancelAndContext.Cancel:\n" + goid.NewStack(ccSkipFrames).Shorts(ccPrepend))
	}
	parl.InvokeCancel(cc.ctx)
}

func (cc cancelAndContext) Context() (ctx context.Context) {
	return cc.ctx
}
