/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plog

import (
	"io"
	"log"
	"os"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

const (
	// frames to skip in public methods
	//	- Log and Info
	//	- public method + doLog() == 2
	//	- indirectly used by Debug GetDebug GetD D
	//	- indirectly used by IsThisDebug IsThisDebugN
	//	- adapted when wrapper functions are used
	logInstDefFrames = 2
	// logInstDebugFrameDelta adjusts to skip only 1 frame
	//	- by adding logInstDebugFrameDelta to LogInstance.stackFramesToSkip
	//	- Used by 4 functions: Debug() GetDebug() D() GetD()
	logInstDebugFrameDelta = -1
	// isThisDebugDelta adjusts to skip only 1 frame
	//	- adding isThisDebugDelta to LogInstance.stackFramesToSkip.
	//	- Used by 2 functions: IsThisDebug() IsThisDebugN()
	isThisDebugDelta = -1
)

// LogInstance provide logging delegating to log.Output
type LogInstance struct {
	isSilence atomic.Bool // when true, [LogInstance.Info] should not print
	isDebug   atomic.Bool // when true, debug is enabled everywhere
	// updated by [LogInstance.SetRegexp]
	//	- used to determine function-level debug
	infoRegexp atomic.Pointer[regexp.Regexp]

	// outLock protects writer and output ensuring thread-safety
	outLock sync.Mutex
	// invoked by Logw printing without ensuring trailing newline
	writer io.Writer
	// output function for writer obtained from [log.New]
	output func(calldepth int, s string) error

	// stackFramesToSkip is used for determining debug status and to get
	// a printable code location.
	// stackFramesToSkip default value is 2, which is one for the invocation of
	// Log() etc., and one for the intermediate doLog() function.
	// Debug(), GetDebug(), D(), GetD(), IsThisDebug() and IsThisDebugN()
	// uses stackFramesToSkip to determine whether the invoking code
	// location has debug active.

	// stackFramesToSkip defaults to logInstDefFrames = 2, which is ?.
	// NewLogFrames(…, extraStackFramesToSkip int) allos to skip
	// additional stack frames
	stackFramesToSkip int
}

// NewLog gets a logger for Fatal and Warning for specific output
func NewLog(writers ...io.Writer) (lg *LogInstance) {
	var writer io.Writer
	if len(writers) > 0 {
		writer = writers[0]
	}
	if writer == nil {
		writer = os.Stderr
	}
	var logger = GetLog(writer)
	if logger == nil {
		logger = log.New(writer, "", 0)
	}
	return &LogInstance{
		writer:            writer,
		output:            logger.Output,
		stackFramesToSkip: logInstDefFrames,
	}
}

// NewLog gets a logger for Fatal and Warning for specific output
func NewLogFrames(writer io.Writer, extraStackFramesToSkip int) (lg *LogInstance) {
	if extraStackFramesToSkip < 0 {
		panic(perrors.Errorf("newLog extraStackFramesToSkip < 0: %d", extraStackFramesToSkip))
	}
	lg = NewLog(writer)
	lg.stackFramesToSkip += extraStackFramesToSkip
	return
}

// Log always prints
//   - if debug is enabled, code location is appended
func (g *LogInstance) Log(format string, a ...interface{}) {
	g.doLog(format, a...)
}

// Logw always prints
//   - Logw does not ensure ending newline
func (g *LogInstance) Logw(format string, a ...interface{}) {
	g.invokeWriter(Sprintf(format, a...))
}

// Info prints unless silence has been configured with SetSilence(true)
//   - IsSilent determines the state of silence
//   - if debug is enabled, code location is appended
func (g *LogInstance) Info(format string, a ...interface{}) {
	if g.isSilence.Load() {
		return
	}
	g.doLog(format, a...)
}

// Debug outputs only if debug is configured globally or for the executing function
//   - code location is appended
func (g *LogInstance) Debug(format string, a ...any) {
	var cloc *pruntime.CodeLocation
	if !g.isDebug.Load() {
		regExp := g.infoRegexp.Load()
		if regExp == nil {
			return // debug: false regexp: nil return: noop
		}
		cloc = pruntime.NewCodeLocation(g.stackFramesToSkip + logInstDebugFrameDelta)
		if !regExp.MatchString(cloc.FuncName) {
			return // debug: false regexp: no match return: noop
		}
	} else {
		cloc = pruntime.NewCodeLocation(g.stackFramesToSkip + logInstDebugFrameDelta)
	}
	g.invokeOutput(pruntime.AppendLocation(Sprintf(format, a...), cloc))
}

// GetDebug returns a function value that can be used to invokes logging
//   - outputs if debug is enabled for the specified caller frame
//   - the caller frame is appended
//   - the function value can be passed around or invoked later
func (g *LogInstance) GetDebug(skipFrames int) (debug func(format string, a ...any)) {

	// code location appended to each log ouput
	var cloc *pruntime.CodeLocation
	var frameNo = g.stackFramesToSkip + logInstDebugFrameDelta + skipFrames
	if frameNo < 1 {
		frameNo = 1
	}
	cloc = pruntime.NewCodeLocation(frameNo)

	// determine if printing should be carried out
	var doPrint = g.isDebug.Load() // global debug
	if !doPrint {
		// check if code location is specified as debug
		var regExp = g.infoRegexp.Load()
		doPrint = regExp != nil && regExp.MatchString(cloc.FuncName)
	}
	if !doPrint {
		return NoPrint // no debug return: no-op function
	}

	return NewOutputInvoker(
		cloc,
		g.invokeOutput,
	).Invoke
}

// GetD returns a function value that always invokes logging
//   - the caller frame is appended
//   - D is meant for temporary output intended to be removed
//     prior to check-in
//   - the function value can be passed around or invoked later
func (g *LogInstance) GetD(skipFrames int) (debug func(format string, a ...interface{})) {
	var frameNo = g.stackFramesToSkip + logInstDebugFrameDelta + skipFrames
	if frameNo < 1 {
		frameNo = 1
	}

	return NewOutputInvoker(
		pruntime.NewCodeLocation(frameNo),
		g.invokeOutput,
	).Invoke
}

// D always prints with code location. Thread-safe
//   - D is meant for temporary output intended to be removed
//     prior to check-in
func (g *LogInstance) D(format string, a ...interface{}) {
	g.invokeOutput(
		pruntime.AppendLocation(
			Sprintf(format, a...),
			pruntime.NewCodeLocation(g.stackFramesToSkip+logInstDebugFrameDelta),
		))
}

// if SetDebug is true, Debug prints everywhere produce output
//   - other printouts have location appended
//   - More selective debug printing can be achieved using SetInfoRegexp
//     that matches on function names.
func (g *LogInstance) SetDebug(debug bool) {
	g.isDebug.Store(debug)
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
func (g *LogInstance) SetRegexp(regExp string) (err error) {

	// compile any provided regexp or vakue is nil
	var regExpPt *regexp.Regexp
	if regExp != "" {
		if regExpPt, err = regexp.Compile(regExp); err != nil {
			return perrors.Errorf("regexp.Compile: %w", err)
		}
	}

	g.infoRegexp.Store(regExpPt)

	return
}

// SetSilent(true) prevents Info() invocations from printing
func (g *LogInstance) SetSilent(silent bool) {
	g.isSilence.Store(silent)
}

// IsThisDebug returns whether the executing code location
// has debug logging enabled
//   - true when -debug globally enabled using SetDebug(true)
//   - true when the -verbose regexp set with SetRegexp matches
func (g *LogInstance) IsThisDebug() (isDebug bool) {
	if g.isDebug.Load() {
		return true // global debug is on return: true
	}
	regExp := g.infoRegexp.Load()
	if regExp == nil {
		return false
	}
	cloc := pruntime.NewCodeLocation(g.stackFramesToSkip + isThisDebugDelta)
	return regExp.MatchString(cloc.FuncName)
}

// IsThisDebugN returns whether the specified stack frame
// has debug logging enabled. 0 means caller of IsThisDebugN.
//   - true when -debug globally enabled using SetDebug(true)
//   - true when the -verbose regexp set with SetRegexp matches
func (g *LogInstance) IsThisDebugN(skipFrames int) (isDebug bool) {
	if isDebug = g.isDebug.Load(); isDebug {
		return // global debug on return: true
	}

	var regExp = g.infoRegexp.Load()
	if regExp == nil {
		return // no regexp return: false
	}
	var cloc = pruntime.NewCodeLocation(g.stackFramesToSkip + isThisDebugDelta + skipFrames)
	isDebug = regExp.MatchString(cloc.FuncName)
	return
}

// IsSilent if true it means that Info does not print
func (g *LogInstance) IsSilent() (isSilent bool) {
	return g.isSilence.Load()
}

// invokeOutput invokes the writer’s output function with mutual exclusion
func (g *LogInstance) invokeOutput(s string) {
	g.outLock.Lock()
	defer g.outLock.Unlock()

	if err := g.output(0, s); err != nil {
		panic(perrors.Errorf("LogInstance output: %w", err))
	}
}

// invokeWriter invokes writer with mutual exclusion
func (g *LogInstance) invokeWriter(s string) {
	g.outLock.Lock()
	defer g.outLock.Unlock()

	if _, err := g.writer.Write([]byte(s)); err != nil {
		panic(perrors.Errorf("LogInstance writer: %w", err))
	}
}

// doLog invokes the writer’s output function for Log and Info
func (g *LogInstance) doLog(format string, a ...interface{}) {
	s := Sprintf(format, a...)
	if g.isDebug.Load() {
		s = pruntime.AppendLocation(s, pruntime.NewCodeLocation(g.stackFramesToSkip))
	}
	g.invokeOutput(s)
}
