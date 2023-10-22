/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plog

import (
	"os"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const pruntimeDframes = 1

// stderrLogLogger is a shared log.Logger instance for stderr.
//   - using this for output ensures thread-safety with plog and parl packages
var stderrLogLogger = GetLog(os.Stderr)

// D prints to stderr with code location. Thread-safe.
//   - D is like parl.D but can be used by packages imported by parl
//   - D is meant for temporary output intended to be removed before check-in
func D(format string, a ...interface{}) {
	if err := stderrLogLogger.Output(0,
		pruntime.AppendLocation(
			Sprintf(format, a...),
			pruntime.NewCodeLocation(pruntimeDframes),
		)); err != nil {
		panic(perrors.ErrorfPF("log.Logger.Output error: %w", err))
	}
}
