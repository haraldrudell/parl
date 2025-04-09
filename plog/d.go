/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plog

import (
	"os"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

// stderr is overriding output that may feature:
//   - buffered output or
//   - status terminal
//   - intended to eventually output to standard error
var stderr atomic.Pointer[func(format string, a ...any)]

// stderrLogLogger is a shared log.Logger instance for stderr.
//   - ensures thread-safety with plog and parl packages
var stderrLogLogger = GetLog(os.Stderr)

// D prints queued to stderr with code location. Thread-safe
//   - replaces [parl.D] when buffered output is in use
//   - buffered output may go via status terminal
func D(format string, a ...any) { D2(format, a...) }

func D2(format string, a ...any) {

	// get string
	var s = pruntime.AppendLocation(
		Sprintf(format, a...),
		pruntime.NewCodeLocation(outDframes),
	)

	// try override
	if logp := stderr.Load(); logp != nil {
		if log := *logp; log != nil {
			log(s)
			return
		}
	}

	// thread-safe output to standrad error
	if err := stderrLogLogger.Output(0, s); err != nil {
		panic(perrors.ErrorfPF("log.Logger.Output error: %w", err))
	}
}

// SwapD sets or gets the current D intercept function
func SwapD(parlLog ...func(format string, a ...any)) (currrentLog func(format string, a ...any)) {

	// set case
	if len(parlLog) > 0 {
		var logFn = parlLog[0]
		if oldp := stderr.Swap(&logFn); oldp != nil {
			currrentLog = *oldp
		}
		return
	}

	// get case
	if p := stderr.Load(); p != nil {
		currrentLog = *p
	}

	return
}

const (
	// stack frames for [D2]
	outDframes = 2
)
