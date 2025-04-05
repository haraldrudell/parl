/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plogt

import (
	"os"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plog"
	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/pterm"
)

// SetD sets the status terminal for debug logging
func SetD(statusTerminal *pterm.StatusTerminal) {
	stStatic.Store(statusTerminal)
}

// stStatic is thread-safe storage for status terminal
var stStatic atomic.Pointer[pterm.StatusTerminal]

// stderrLogLogger is a shared log.Logger instance for stderr.
//   - using this for output ensures thread-safety with plog and parl packages
var stderrLogLogger = plog.GetLog(os.Stderr)

// D provides [parl.D] that may be directed to status terminal
func D(format string, a ...any) {

	// get output string with code location appended
	var s = pruntime.AppendLocation(
		plog.Sprintf(format, a...),
		pruntime.NewCodeLocation(plogtDSkipFrames),
	)

	// log to statusTerminal if present
	var statusTerminal = stStatic.Load()
	if statusTerminal != nil {
		statusTerminal.Log(s)
		return
	}

	// log to standard error, thread-safe
	if err := stderrLogLogger.Output(0, s); err != nil {
		panic(perrors.ErrorfPF("log.Logger.Output error: %w", err))
	}
}

// D provides [parl.Log] that may be directed to status terminal
func Log(format string, a ...any) {

	// get output string
	var s = plog.Sprintf(format, a...)

	// log to statusTerminal if present
	var statusTerminal = stStatic.Load()
	if statusTerminal != nil {
		statusTerminal.LogTimeStamp(s)
		return
	}

	// log to standard error, thread-safe
	if err := stderrLogLogger.Output(0, s); err != nil {
		panic(perrors.ErrorfPF("log.Logger.Output error: %w", err))
	}
}

const (
	plogtDSkipFrames = 1
)
