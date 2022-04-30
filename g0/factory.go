/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

var GoGroupFactory parl.GoGroupFactory = &goCreatorFactory{}

type goCreatorFactory struct{}

var _ parl.GoGroupFactory = &goCreatorFactory{}

func (gf *goCreatorFactory) NewGoGroup(ctx context.Context) (goCreator parl.GoGroup) {
	return NewGoGroup(ctx)
}

var GoerGroupFactory parl.GoerGroupFactory = &goerFactory{}

type goerFactory struct{}

func (gf *goerFactory) NewGoerGroup(ctx context.Context) (goer parl.GoerGroup) {
	return NewGoerGroup(ctx)
}
