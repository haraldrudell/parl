/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"

	"github.com/haraldrudell/parl/perrors"
)

// CancelContextDo implements parl.CancelContext.
// CancelContextDo is a context with a thread-safe, race-free,
// idempotent Cancel method
type CancelContextDo struct {
	context.Context
	cancel context.CancelFunc
}

// NewCancelContext instantiates parl.CancelContext, a context
// with Cancel exposed as a method
func NewCancelContext(ctx context.Context) (cancelCtx CancelContext) {
	c := CancelContextDo{}
	c.Context, c.cancel = context.WithCancel(ctx)
	return
}

// Cancel cancels this context
func (cc *CancelContextDo) Cancel() {
	if cc == nil || cc.cancel == nil {
		var s string
		if cc == nil {
			s = "context"
		} else {
			s = "context.cancel"
		}
		panic(perrors.Errorf("CancelContextDo.Cancel with %s nil", s))
	}
	cc.cancel()
}
