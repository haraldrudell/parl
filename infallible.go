/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

// Infallible is an error sink logging to standard error
//   - intended for failure recovery of threads that should not fail
//   - use should be limited to threads reading from sockets that cannot be terminated,
//     ie. standard input or the udev netlink socket
//   - outputs the error with stack trace and the stack trace invoking [Infallible.AddError]
var Infallible ErrorSink1 = &infallible{}

type infallible struct{}

func (i *infallible) AddError(err error) {
	Log("\nInfallible FAILED\n\nerr:\n%s\n\nInfallible invocation:\n%s\n",
		perrors.Long(err), pruntime.DebugStack(0))
}
