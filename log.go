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

// stderrLogger prints to standard error with proper code location
var stderrLogger = plog.NewLogFrames(nil, logStackFramesToSkip)

// stderrLogger prints to standard out with proper code location
var stdoutLogger = plog.NewLogFrames(os.Stdout, logStackFramesToSkip)

// Out always prints to standard out
//   - if debug is enabled, code location is appended
func Out(format string, a ...interface{}) {
	stdoutLogger.Log(format, a...)
}

// Outw always prints to standard out without ensuring terminating newline
func Outw(format string, a ...interface{}) {
	stdoutLogger.Logw(format, a...)
}

// Log always prints to standard error
//   - if debug is enabled, code location is appended
func Log(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}

// Logw always prints to standard error
//   - Logw outputs without ensruing ending newline
func Logw(format string, a ...interface{}) {
	stderrLogger.Logw(format, a...)
}

// Console always prints to interactive standard out
//   - intended for command-line interactivity.
//   - if debug is enabled, code location is appended.
func Console(format string, a ...interface{}) {
	stderrLogger.Log(format, a...)
}

// Consolew always prints to standard out
//   - intended for command-line interactivity.
//   - Consolew does not ensure ending newline
func Consolew(format string, a ...interface{}) {
	stderrLogger.Logw(format, a...)
}

// Info prints unless silence has been configured with SetSilence(true)
//   - Info outputs to standard error
//   - IsSilent() deteremines the state of silence
//   - if debug is enabled, code location is appended
func Info(format string, a ...interface{}) {
	stderrLogger.Info(format, a...)
}

// Debug outputs only if debug is configured globally or for the executing function
//   - Debug outputs to standard error
//   - code location is appended
func Debug(format string, a ...interface{}) {
	stderrLogger.Debug(format, a...)
}

// GetDebug returns a function value that can be used to invokes logging
//   - output to stderr if debug is enabled for the specified caller frame
//   - the caller frame is appended
//   - the function value can be passed around or invoked later
func GetDebug(skipFrames int) (debug func(format string, a ...interface{})) {
	return stderrLogger.GetDebug(skipFrames)
}

// GetD returns a function value that always invokes logging
//   - output to stderr
//   - the caller frame is appended
//   - D is meant for temporary output intended to be removed
//     prior to check-in
//   - the function value can be passed around or invoked later
func GetD(skipFrames int) (debug func(format string, a ...interface{})) {
	return stderrLogger.GetD(skipFrames)
}

// IsThisDebug returns whether the executing code location
// has debug logging enabled
//   - true when -debug globally enabled using SetDebug(true)
//   - true when the -verbose regexp set with SetRegexp matches
//   - matched against [pruntime.CodeLocation.FuncName]
//   - “github.com/haraldrudell/parl/mains.(*Executable).AddErr”
func IsThisDebug() bool {
	return stderrLogger.IsThisDebug()
}

// IsThisDebugN returns whether the specified stack frame
// has debug logging enabled. 0 means caller of IsThisDebugN.
//   - true when -debug globally enabled using SetDebug(true)
//   - true when the -verbose regexp set with SetRegexp matches
func IsThisDebugN(skipFrames int) (isDebug bool) {
	return stderrLogger.IsThisDebugN(skipFrames)
}

// IsSilent if true it means that Info does not print
func IsSilent() (isSilent bool) {
	return stderrLogger.IsSilent()
}

// SetRegexp defines a regular expression for function-level debug
// printing to stderr.
//   - SetRegexp affects Debug() GetDebug() IsThisDebug() IsThisDebugN()
//     functions.
//
// # Regular Expression
//
// Regular expression is the RE2 [syntax] used by golang.
// command-line documentation: “go doc regexp/syntax”.
// The regular expression is matched against code location.
//
// # Code Location Format
//
// Code location is the fully qualified function name
// for the executing code line being evaluated.
// This is a fully qualified golang package path, ".",
// a possible type name in parenthesis ending with "." and the function name.
//
//   - method with pointer receiver:
//   - — "github.com/haraldrudell/parl/mains.(*Executable).AddErr"
//   - — sample regexp: mains...Executable..AddErr
//   - top-level function:
//   - — "github.com/haraldrudell/parl/g0.NewGoGroup"
//   - — sample regexp: g0.NewGoGroup
//
// To obtain the fully qualified function name for a particular location:
//
//	parl.Log(pruntime.NewCodeLocation(0).String())
//
// [syntax]: https://github.com/google/re2/wiki/Syntax
func SetRegexp(regExp string) (err error) {
	return stderrLogger.SetRegexp(regExp)
}

// SetSilent(true) prevents Info() invocations from printing
func SetSilent(silent bool) {
	stderrLogger.SetSilent(silent)
}

// if SetDebug is true, Debug prints everywhere produce output
//   - other printouts have location appended
//   - More selective debug printing can be achieved using SetInfoRegexp
//     that matches on function names.
func SetDebug(debug bool) {
	stderrLogger.SetDebug(debug)
}

// D always prints to stderr with code location. Thread-safe
//   - D is meant for temporary output intended to be removed
//     prior to check-in
func D(format string, a ...interface{}) {

	// check for intercept
	if plog.SwapD() != nil {
		plog.D2(format, a...)
		return
	}

	stderrLogger.D(format, a...)
}
