/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/pdebug"
)

const (
	g1ccSkipFrames = 1
	g1ccPrepend    = "— "
)

// goContext is a promotable private field
//   - public methods: Cancel() Context() EntityID()
//   - goContext is based on parl.NewCancelContext
type goContext struct {
	goEntityID // EntityID()
	wg         parl.WaitGroupCh
	//	- updatable, therefore must be atomic access
	ctxp atomic.Pointer[context.Context]
	// cancelListener is function invoked immediately prior to
	// parl.InvokeCancel
	cancelListener atomic.Pointer[func()]
}

// newGoContext returns a subordinate context with Cancel and Context methods
//   - uses non-pointer atomics
func newGoContext(fieldp *goContext, ctx context.Context) (g *goContext) {
	if ctx == nil {
		panic(parl.NilError("ctx"))
	}
	if fieldp != nil {
		g = fieldp
		*g = goContext{goEntityID: *newGoEntityID()}
	} else {
		g = &goContext{goEntityID: *newGoEntityID()}
	}
	var ctx2 = parl.NewCancelContext(ctx)
	g.ctxp.Store(&ctx2)
	return
}

// Cancel signals shutdown to all threads of a thread-group.
func (c *goContext) Cancel() {
	if f := c.cancelListener.Load(); f != nil {
		(*f)()
	}
	// if caller is debug, debug-print cancel action
	if parl.IsThisDebugN(g1ccSkipFrames) {
		parl.GetDebug(g1ccSkipFrames)("CancelAndContext.Cancel:\n" + pdebug.NewStack(g1ccSkipFrames).Shorts(g1ccPrepend))
	}
	parl.InvokeCancel(*c.ctxp.Load())
}

// Context returns the context of this cancelAndContext.
//   - Context is used to detect cancel using the receive channel Context.Done.
//   - Context cancellation has happened when Context.Err is non-nil.
func (c *goContext) Context() (ctx context.Context) {
	return *c.ctxp.Load()
}

// addNotifier adds a stack trace to every Cancel invocation
//
// Usage:
//
//	threadGroup := g0.NewGoGroup(ctx)
//	threadGroup.(*g0.GoGroup).addNotifier(func(slice pruntime.StackSlice) {
//	  parl.D("CANCEL %s %s\n\n\n\n\n", g0.GoChain(threadGroup), slice)
//	})
// func (c *goContext) addNotifier(notifier func(slice pruntime.StackSlice)) {
// 	for {
// 		var ctxp0 = c.ctxp.Load()
// 		var ctx = parl.AddNotifier1(*ctxp0, notifier)
// 		if c.ctxp.CompareAndSwap(ctxp0, &ctx) {
// 			return
// 		}
// 	}
// }

func (c *goContext) setCancelListener(f func()) {
	c.cancelListener.Store(&f)
}
