/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import "github.com/haraldrudell/parl/breakcycle"

var _ = func() (i int) {
	breakcycle.PtimeExport(Short)
	return
}()
