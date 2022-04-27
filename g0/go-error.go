/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type GoErrorDo struct {
	error
	source parl.GoErrorSource
	goer   parl.Goer
}

func NewGoError(err error, source parl.GoErrorSource, goer parl.Goer) (goError parl.GoError) {
	return &GoErrorDo{
		error:  err,
		source: source,
		goer:   goer,
	}
}

func (ge *GoErrorDo) Source() (source parl.GoErrorSource) {
	return ge.source
}

func (ge *GoErrorDo) GetError() (err error) {
	return ge.error
}

func (ge *GoErrorDo) Goer() (goer parl.Goer) {
	return ge.goer
}

func (ge *GoErrorDo) String() (s string) {
	return ge.source.String() +
		"\x20" + perrors.Short(ge.error) +
		"\x20" + ge.goer.String()
}
