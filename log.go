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
	// extra stack frame for the parl indirection, used by 6 functions:
	// Debug() GetDebug() D() GetD() IsThisDebug() IsThisDebugN()
	logStackFramesToSkip = 1
)

var stderrLogger = plog.NewLogFrames(nil, logStackFramesToSkip)
var stdoutLogger = plog.NewLogFrames(os.Stdout, logStackFramesToSkip)

// Out always prints payload output to stdout.
func Out(format string, a ...interface{}) {
	stdoutLogger.Log(format, a...)
}

// Outw always prints payload output to stdout without terminating newline.
func Outw(format string, a ...interface{}) {
	stderrLogger.Logw(format, a...)
}

// Log invocations always print and output to stderr.
// if debug is enabled, code location is appended.
func Log(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}

// Logw invocations always print and output to stderr.
// Logw outputs without appending newline.
func Logw(format string, a ...interface{}) {
	stderrLogger.Logw(format, a...)
}

// Console always prints to stdout, intended for command-line interactivity.
// if debug is enabled, code location is appended.
func Console(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}

// Consolew always prints to stdout intended for command-line interactivity.
// Consolew does not append a newline.
func Consolew(format string, a ...interface{}) {
	stderrLogger.Logw(format, a...)
}

// Info prints unless silence has been configured with SetSilence(true).
// Info outputs to stderr.
// IsSilent() deteremines the state of silence.
// if debug is enabled, code location is appended.
func Info(format string, a ...interface{}) {
	stderrLogger.Info(format, a...)
}

// Debug outputs only if debug is configured or the code location package
// matches regexp. Debug outputs to stderr and code location is appended.
func Debug(format string, a ...interface{}) {
	stderrLogger.Debug(format, a...)
}

// GetDebug obtains a Debug function based on the invocation location
// adjusted by skipFrames.
// The return function is used for later execution.
// The returned function outputs only if debug is configured or
// the code location package matches regexp.
// The returned function outputs to stderr and code location is appended.
func GetDebug(skipFrames int) (debug func(format string, a ...interface{})) {
	return stderrLogger.GetDebug(skipFrames)
}

// GetD obtains always printing D based on the invocation location for later execution
func GetD(skipFrames int) (debug func(format string, a ...interface{})) {
	return stderrLogger.GetD(skipFrames)
}

// IsThisDebug returns true when debug is globally set using Debug(true) or
// when debug logging is configured for the code location using SetRegexp().
func IsThisDebug() bool {
	return stderrLogger.IsThisDebug()
}

// IsThisDebugN returns true when debug is globally set using Debug(true) or
// when debug logging is configured for the code location adjusted by skipFrames
// using SetRegexp().
func IsThisDebugN(skipFrames int) (isDebug bool) {
	return stderrLogger.IsThisDebugN(skipFrames)
}

// IsSilent if true it means that Info does not print
func IsSilent() (isSilent bool) {
	return stderrLogger.IsSilent()
}

// SetRegexp defines a regular expression for selective debug
// printing to stderr.
// SetRegexp affects Debug() GetDebug() IsThisDebug() IsThisDebugN()
// functions.
//
// # Regular Expression
//
// Regular expression is the RE2 [syntax] used by golang.
// command-line documentation: go doc regexp/syntax.
// The string the regular expression is matched against is a fully
// qualified function name, ie.
//
// # Code Location Format
//
// The regular expression is matched agains the fully qualified function name
// for the code line being evaluated.
// This is a fully qualified golang package path, ".",
// a possible type name in parenthesis ending with "." and the function name.
//
// "github.com/haraldrudell/parl/mains.(*Executable).AddErr"
//
// To obtain the fully qualified function name for a particular location:
//
//	parl.Log(pruntime.NewCodeLocation(0).String())
//
// [syntax]: https://github.com/google/re2/wiki/Syntax
func SetRegexp(regExp string) (err error) {
	return stderrLogger.SetRegexp(regExp)
}

// SetSilent(true) prevents Info() invocations from printing.
func SetSilent(silent bool) {
	stderrLogger.SetSilent(silent)
}

// if SetDebug is true, all Debug prints everywhere produce output.
// More selective debug printing can be achieved using SetInfoRegexp
// that matches package names.
func SetDebug(debug bool) {
	stderrLogger.SetDebug(debug)
}

// D always prints to stderr with code location and is thread safe.
// D is meant for temporary output using invocations that are removed
// prior to source code repository check-in.
func D(format string, a ...interface{}) {
	stderrLogger.D(format, a...)
}
