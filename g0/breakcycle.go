/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"time"

	"github.com/haraldrudell/parl/breakcycle"
	"github.com/haraldrudell/parl/perrors"
)

var short func(tim ...time.Time) (s string)

var _ = func() (i int) {
	breakcycle.G0Import(setShort)
	return
}()

func setShort(v interface{}) {
	var ok bool
	if short, ok = v.(func(tim ...time.Time) (s string)); !ok {
		panic(perrors.Errorf("setShort: v bad type: %T %[1]v", v))
	}
}
