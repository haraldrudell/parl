/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"github.com/haraldrudell/parl"
)

type GoerRuntime struct {
	wg parl.WaitGroup
	g0 parl.Go
}
