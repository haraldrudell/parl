/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

// cancelAndContext makes parl.CancelAndContext private
type cancelAndContext struct {
	parl.CancelAndContext // Cancel() Context()
}

func newCancelAndContext(ctx context.Context) (cc *cancelAndContext) {
	return &cancelAndContext{CancelAndContext: *parl.NewCancelAndContext(ctx)}
}
