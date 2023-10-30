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

func (f *goGroupFactory) NewGoGroup(ctx context.Context, onFirstFatal ...parl.GoFatalCallback) (goGroup parl.GoGroup) {
	return NewGoGroup(ctx, onFirstFatal...)
}
