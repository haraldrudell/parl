/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"os"

	"github.com/haraldrudell/parl/plog"
)

const (
	logStackFramesToSkip = 1
)

var stderrLogger = plog.NewLogFrames(nil, logStackFramesToSkip)
var stdoutLogger = plog.NewLogFrames(os.Stdout, logStackFramesToSkip)

// Out prints extected output to stdout
func Out(format string, a ...interface{}) {
	stdoutLogger.Log(format, a...)
}

func Outw(format string, a ...interface{}) {
	stderrLogger.Logw(format, a...)
}

// Log invocations always print
// if debug is enabled, code location is appended
func Log(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}

// Logw invocations always print.
// Logw outputs withoput appending newline
func Logw(format string, a ...interface{}) {
	stderrLogger.Logw(format, a...)
}

// Console always print intended for command-line interactivity
// if debug is enabled, code location is appended
func Console(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}

// Consolew always print intended for command-line interactivity
// Consolew does not append a newline
func Consolew(format string, a ...interface{}) {
	stderrLogger.Logw(format, a...)
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

// GetDebug obtains a Debug based on the invocation location for later execution
func GetDebug(skipFrames int) (debug func(format string, a ...interface{})) {
	return stderrLogger.GetDebug(skipFrames)
}

// IsThisDebug returns whether debug logging is configured for the executing function
func IsThisDebug() bool {
	return stderrLogger.IsThisDebug()
}

func IsThisDebugN(skipFrames int) (isDebug bool) {
	return stderrLogger.IsThisDebugN(skipFrames)
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
