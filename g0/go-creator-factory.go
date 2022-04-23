/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"

	"github.com/haraldrudell/parl"
)

var GoCreatorFy = &goCreatorFactory{}

type goCreatorFactory struct{}

var _ parl.GoCreatorFactory = &goCreatorFactory{}

func (gf *goCreatorFactory) NewGoCreator(ctx context.Context) (goCreator parl.GoCreator) {
	return NewGoCreator(ctx)
}
