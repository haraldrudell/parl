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
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	g1ccSkipFrames = 1
	g1ccPrepend    = "— "
)

// goContext is a promotable private field with Cancel and Context methods only.
//   - goContext is based on parl.NewCancelContext
type goContext struct {
	goEntityID
	c              atomic.Pointer[context.Context]
	cancelListener func()
}

// newGoContext returns a subordinate context with Cancel and Context methods
func newGoContext(ctx context.Context) (gc *goContext) {
	if ctx == nil {
		panic(perrors.NewPF("ctx cannot be nil"))
	}
	var ctx2 = parl.NewCancelContext(ctx)
	c := goContext{
		goEntityID: *newGoEntityID(),
	}
	c.c.Store(&ctx2)
	return &c
}

// AddNotifier adds a stack trace to every Cancel invocation
//
// Usage:
//
//	threadGroup := g0.NewGoGroup(ctx)
//	threadGroup.(*g0.GoGroup).AddNotifier(func(slice pruntime.StackSlice) {
//	  parl.D("CANCEL %s %s\n\n\n\n\n", g0.GoChain(threadGroup), slice)
//	})
func (c *goContext) AddNotifier(notifier func(slice pruntime.StackSlice)) {
	var ctx = parl.AddNotifier1(*c.c.Load(), notifier)
	c.c.Store(&ctx)
}

// Cancel signals shutdown to all threads of a thread-group.
func (c *goContext) Cancel() {
	if f := c.cancelListener; f != nil {
		f()
	}
	// if caller is debug, debug-print cancel action
	if parl.IsThisDebugN(g1ccSkipFrames) {
		parl.GetDebug(g1ccSkipFrames)("CancelAndContext.Cancel:\n" + pdebug.NewStack(g1ccSkipFrames).Shorts(g1ccPrepend))
	}
	parl.InvokeCancel(*c.c.Load())
}

// Context returns the context of this cancelAndContext.
//   - Context is used to detect cancel using the receive channel Context.Done.
//   - Context cancellation has happened when Context.Err is non-nil.
func (c *goContext) Context() (ctx context.Context) {
	return *c.c.Load()
}
