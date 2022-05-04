/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package goid

import "github.com/haraldrudell/parl/breakcycle"

var _ = func() (i int) {
	breakcycle.SetV(NewStack)
	return
}()
