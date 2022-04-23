/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

func Infallible(err error) {
	Log("\nInfallible FAILED\n\nerr:\n%s\n\nInfallible invocation:\n%s\n",
		perrors.Long(err), pruntime.DebugStack(0))
}
