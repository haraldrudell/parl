/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

var GoGroupFactory parl.GoFactory = &goGroupFactory{}

type goGroupFactory struct{}

func (gf *goGroupFactory) NewGoGroup(ctx context.Context, onFirstFatal ...parl.GoFatalCallback) (g1 parl.GoGroup) {
	return NewGoGroup(ctx, onFirstFatal...)
}
