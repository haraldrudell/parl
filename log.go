/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"os"

	"github.com/haraldrudell/parl/parlay"
)

const (
	logStackFramesToSkip = 1
)

var stderrLogger = parlay.NewLogFrames(nil, logStackFramesToSkip)
var stdoutOutput = parlay.GetLog(os.Stdout).Output

// Out prints extected output to stdout
func Out(format string, a ...interface{}) {
	stdoutOutput(0, Sprintf(format, a))
}

// Log invocations always print
// if debug is enabled, code location is appended
func Log(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}

// Console always print intended for command-line interactivity
// if debug is enabled, code location is appended
func Console(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}

// Info prints unless silence has been configured with SetSilence(true)
// IsSilent deteremines the state of silence
// if debug is enabled, code location is appended
func Info(format string, a ...interface{}) {
	stderrLogger.Info(format, a...)
}

// Debug outputs only if debug is configured or the code location package matches regexp
func Debug(format string, a ...interface{}) {
	stderrLogger.Debug(format, a...)
}

// IsThisDebug returns whether debug logging is configured for the executing function
func IsThisDebug() bool {
	return stderrLogger.IsThisDebug()
}

// IsSilent if true it means that Info does not print
func IsSilent() (isSilent bool) {
	return stderrLogger.IsSilent()
}

func SetRegexp(regExp string) (err error) {
	return stderrLogger.SetRegexp(regExp)
}

// SetSilent
func SetSilent(silent bool) {
	stderrLogger.SetSilent(silent)
}

// if SetDebug is true, all Debug prints everywhere produce output.
// More selective debug printing can be achieved using SetInfoRegexp that matches package names.
func SetDebug(debug bool) {
	stderrLogger.SetDebug(debug)
}

// D prints to stderr with code location
// Thread safe. D is meant for temporary output intended to be removed before check-in
func D(format string, a ...interface{}) {
	stderrLogger.D(format, a...)
}
