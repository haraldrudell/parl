/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pdebug"
	"github.com/haraldrudell/parl/perrors"
)

const (
	g1ccSkipFrames = 1
	g1ccPrepend    = "— "
)

// goContext is a promotable private field with Cancel and Context methods only.
//   - goContext is based on parl.NewCancelContext
type goContext struct {
	ctx context.Context
}

// newGoContext returns a subordinate context with Cancel and Context methods
func newGoContext(ctx context.Context) (ctx2 *goContext) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	}
	return &goContext{ctx: parl.NewCancelContext(ctx)}
}

// Cancel signals shutdown to all threads of a thread-group.
func (cc *goContext) Cancel() {

	// if caller is debug, debug-print cancel action
	if parl.IsThisDebugN(g1ccSkipFrames) {
		parl.GetDebug(g1ccSkipFrames)("CancelAndContext.Cancel:\n" + pdebug.NewStack(g1ccSkipFrames).Shorts(g1ccPrepend))
	}
	parl.InvokeCancel(cc.ctx)
}

// Context returns the context of this cancelAndContext.
//   - Context is used to detect cancel using the receive channel Context.Done.
//   - Context cancellation has happened when Context.Err is non-nil.
func (cc *goContext) Context() (ctx context.Context) {
	return cc.ctx
}
