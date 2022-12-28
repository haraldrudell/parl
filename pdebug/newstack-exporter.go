/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pdebug

import "github.com/haraldrudell/parl"

var _ = func() (b bool) {
	parl.ImportNewStack(NewStack)
	return
}()
