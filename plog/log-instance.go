/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// plog provides log instances with Log Logw Info Debug D SetDebug SetRegexp
// SetRegexp SetSilent IsThisDebug IsSilent.
// Static delegasting functions are available in parl, like parl.Log.
package plog

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/plogger"
	"github.com/haraldrudell/parl/pruntime"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	logInstDefFrames = 2 // frames to skip in doLog() [Log() + doLog() == 2]
	// logInstDebugFrameDelta adjust stack frame calculation to skip only 1 frame
	// by adding logInstDebugFrameDelta to LogInstance.stackFramesToSkip.
	// Used by 4 functions: Debug() GetDebug() D() GetD()
	logInstDebugFrameDelta = -1
	// isThisDebugDelta adjust stack frame calculation to skip only 1 frame
	// by adding isThisDebugDelta to LogInstance.stackFramesToSkip.
	// Used by 2 functions: IsThisDebug() IsThisDebugN()
	isThisDebugDelta        = -1
	uint32True       uint32 = 1 // because atomic is 32-bit, this is the value used indicating true
)

// LogInstance provide logging delegating to log.Output
type LogInstance struct {
	isSilence  uint32 // atomic
	isDebug    uint32 // atomic
	infoLock   sync.RWMutex
	infoRegexp *regexp.Regexp // behing infoLock
	outLock    sync.Mutex
	writer     io.Writer
	output     func(calldepth int, s string) error
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

var messageSprintf = message.NewPrinter(language.English).Sprintf

func sprintf(format string, a ...interface{}) (s string) {
	if len(a) > 0 {
		s = messageSprintf(format, a...)
	} else {
		s = format
	}
	return
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
	logger := plogger.GetLog(writer)
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

// Log invocations always print
// if debug is enabled, code location is appended
func (lg *LogInstance) Log(format string, a ...interface{}) {
	lg.doLog(format, a...)
}

// Logw invocations always print.
// Logw outputs withoput appending newline
func (lg *LogInstance) Logw(format string, a ...interface{}) {
	lg.invokeWriter(sprintf(format, a...))
}

// Info prints unless silence has been configured with SetSilence(true)
// IsSilent deteremines the state of silence
// if debug is enabled, code location is appended
func (lg *LogInstance) Info(format string, a ...interface{}) {
	if atomic.LoadUint32(&lg.isSilence) != 0 {
		return
	}
	lg.doLog(format, a...)
}

// Debug outputs only if debug is configured or the code location package matches regexp
func (lg *LogInstance) Debug(format string, a ...interface{}) {
	var cloc *pruntime.CodeLocation
	if atomic.LoadUint32(&lg.isDebug) == 0 {
		regExp := lg.getRegexp()
		if regExp == nil {
			return // debug: false regexp: nil
		}
		cloc = pruntime.NewCodeLocation(lg.stackFramesToSkip + logInstDebugFrameDelta)
		if !regExp.MatchString(cloc.FuncName) {
			return // debug: false regexp: no match
		}
	} else {
		cloc = pruntime.NewCodeLocation(lg.stackFramesToSkip + logInstDebugFrameDelta)
	}
	lg.invokeOutput(appendLocation(sprintf(format, a...), cloc))
}

// Debug outputs only if debug is configured or the code location package matches regexp
func (lg *LogInstance) GetDebug(skipFrames int) (debug func(format string, a ...interface{})) {

	// determine if printing
	doPrint := atomic.LoadUint32(&lg.isDebug) != 0
	cloc := pruntime.NewCodeLocation(lg.stackFramesToSkip + logInstDebugFrameDelta + skipFrames)
	if !doPrint {
		regExp := lg.getRegexp()
		doPrint = regExp != nil && regExp.MatchString(cloc.FuncName)
	}

	if !doPrint {
		return NoPrint
	}

	return func(format string, a ...interface{}) {
		lg.invokeOutput(appendLocation(sprintf(format, a...), cloc))
	}
}

func (lg *LogInstance) GetD(skipFrames int) (debug func(format string, a ...interface{})) {
	cloc := pruntime.NewCodeLocation(lg.stackFramesToSkip + logInstDebugFrameDelta + skipFrames)

	return func(format string, a ...interface{}) {
		lg.invokeOutput(appendLocation(sprintf(format, a...), cloc))
	}
}

// D prints to stderr with code location
// Thread safe. D is meant for temporary output intended to be removed before check-in
func (lg *LogInstance) D(format string, a ...interface{}) {
	lg.invokeOutput(appendLocation(sprintf(format, a...), pruntime.NewCodeLocation(lg.stackFramesToSkip+logInstDebugFrameDelta)))
}

// SetDebug prints everything with code location: IsInfo
func (lg *LogInstance) SetDebug(debug bool) {
	var v uint32
	if debug {
		v = uint32True
	}
	atomic.StoreUint32(&lg.isDebug, v)
}

func (lg *LogInstance) SetRegexp(regExp string) (err error) {
	var regExpPt *regexp.Regexp
	if regExp != "" {
		if regExpPt, err = regexp.Compile(regExp); err != nil {
			return perrors.Errorf("regexp.Compile: %w", err)
		}
	}
	lg.infoLock.Lock()
	defer lg.infoLock.Unlock()
	lg.infoRegexp = regExpPt
	return
}

// SetSilent only prints Log
func (lg *LogInstance) SetSilent(silent bool) {
	var v uint32
	if silent {
		v = 1
	}
	atomic.StoreUint32(&lg.isSilence, v)
}

// IsThisDebug returns whether debug logging is configured
func (lg *LogInstance) IsThisDebug() (isDebug bool) {
	if atomic.LoadUint32(&lg.isDebug) != 0 {
		return true
	}
	regExp := lg.getRegexp()
	if regExp == nil {
		return false
	}
	cloc := pruntime.NewCodeLocation(lg.stackFramesToSkip + isThisDebugDelta)
	return regExp.MatchString(cloc.FuncName)
}

func (lg *LogInstance) IsThisDebugN(skipFrames int) (isDebug bool) {
	if atomic.LoadUint32(&lg.isDebug) != 0 {
		return true
	}
	regExp := lg.getRegexp()
	if regExp == nil {
		return false
	}
	cloc := pruntime.NewCodeLocation(lg.stackFramesToSkip + isThisDebugDelta + skipFrames)
	return regExp.MatchString(cloc.FuncName)
}

// IsSilent if true it means that Info does not print
func (lg *LogInstance) IsSilent() (isSilent bool) {
	return atomic.LoadUint32(&lg.isSilence) != 0
}

func (lg *LogInstance) invokeOutput(s string) {
	lg.outLock.Lock()
	defer lg.outLock.Unlock()

	if err := lg.output(0, s); err != nil {
		panic(perrors.Errorf("LogInstance output: %w", err))
	}
}

func (lg *LogInstance) invokeWriter(s string) {
	lg.outLock.Lock()
	defer lg.outLock.Unlock()

	if _, err := lg.writer.Write([]byte(s)); err != nil {
		panic(perrors.Errorf("LogInstance writer: %w", err))
	}
}

func (lg *LogInstance) getRegexp() *regexp.Regexp {
	lg.infoLock.RLock()
	defer lg.infoLock.RUnlock()
	return lg.infoRegexp
}

func (lg *LogInstance) doLog(format string, a ...interface{}) {
	s := sprintf(format, a...)
	if atomic.LoadUint32(&lg.isDebug) != 0 {
		s = appendLocation(s, pruntime.NewCodeLocation(lg.stackFramesToSkip))
	}
	lg.invokeOutput(s)
}

func appendLocation(s string, location *pruntime.CodeLocation) string {
	// insert code location before a possible ending newline
	sNewline := ""
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
		sNewline = "\n"
	}
	return fmt.Sprintf("%s %s%s", s, location.Short(), sNewline)
}
