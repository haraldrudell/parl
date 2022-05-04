/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

var GoGroupFactory parl.GoGroupFactory = &goGroupFactory{}

type goGroupFactory struct{}

var _ parl.GoGroupFactory = &goGroupFactory{}

func (gf *goGroupFactory) NewGoGroup(ctx context.Context) (goCreator parl.GoGroup) {
	return NewGoGroup(ctx)
}

var GoerFactory parl.GoerFactory = &goerFactory{}

type goerFactory struct{}

func (gf *goerFactory) NewGoer(ctx context.Context) (goer parl.Goer) {
	return NewGoerGroup(ctx)
}
