/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

var GoCreatorFy parl.GoCreatorFactory = &goCreatorFactory{}

type goCreatorFactory struct{}

var _ parl.GoCreatorFactory = &goCreatorFactory{}

func (gf *goCreatorFactory) NewGoCreator(ctx context.Context) (goCreator parl.GoGroup) {
	return NewGoCreator(ctx)
}

var GoerFactory parl.GoerFactory = &goerFactory{}

type goerFactory struct{}

func (gf *goerFactory) NewGoer(ctx context.Context) (goer parl.SubGoer) {
	return NewGoer1(ctx)
}
