/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl/error116"
	"github.com/haraldrudell/parl/runt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	logInstDefFrames       = 2
	logInstDebugFrameDelta = -1
	isThisDebugDelta       = -1
)

// LogInstance provide logging delegating to log.Output
type LogInstance struct {
	isSilence         uint32 // atomic
	isDebug           uint32 // atomic
	infoLock          sync.RWMutex
	infoRegexp        *regexp.Regexp // behing infoLock
	output            func(calldepth int, s string) error
	stackFramesToSkip int
}

var sprintf = message.NewPrinter(language.English).Sprintf

// NewLog gets a logger for Fatal and Warning for specific output
func NewLog(writers ...io.Writer) (lg *LogInstance) {
	var writer io.Writer
	if len(writers) > 0 {
		writer = writers[0]
	}
	if writer == nil {
		writer = os.Stderr
	}
	logger := logMap[writer]
	if logger == nil {
		logger = log.New(writer, "", 0)
	}
	return &LogInstance{
		output:            logger.Output,
		stackFramesToSkip: logInstDefFrames,
	}
}

// NewLog gets a logger for Fatal and Warning for specific output
func NewLogFrames(writer io.Writer, extraStackFramesToSkip int) (lg *LogInstance) {
	if extraStackFramesToSkip < 0 {
		panic(error116.Errorf("newLog extraStackFramesToSkip < 0: %d", extraStackFramesToSkip))
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
	var cloc *runt.CodeLocation
	if atomic.LoadUint32(&lg.isDebug) == 0 {
		regExp := lg.getRegexp()
		if regExp == nil {
			return // debug: false regexp: nil
		}
		cloc = runt.NewCodeLocation(lg.stackFramesToSkip + logInstDebugFrameDelta)
		if !regExp.MatchString(cloc.FuncName) {
			return // debug: false regexp: no match
		}
	} else {
		cloc = runt.NewCodeLocation(lg.stackFramesToSkip + logInstDebugFrameDelta)
	}
	if err := lg.output(0, appendLocation(sprintf(format, a...), cloc)); err != nil {
		panic(error116.Errorf("LogInstance output: %w", err))
	}
}

// D prints to stderr with code location
// Thread safe. D is meant for temporary output intended to be removed before check-in
func (lg *LogInstance) D(format string, a ...interface{}) {
	if err := lg.output(0, appendLocation(sprintf(format, a...), runt.NewCodeLocation(lg.stackFramesToSkip+logInstDebugFrameDelta))); err != nil {
		panic(error116.Errorf("LogInstance output: %w", err))
	}
}

// SetDebug prints everything with code location: IsInfo
func (lg *LogInstance) SetDebug(debug bool) {
	var v uint32
	if debug {
		v = 1
	}
	atomic.StoreUint32(&lg.isDebug, v)
}

func (lg *LogInstance) SetRegexp(regExp string) (err error) {
	var regExpPt *regexp.Regexp
	if regExp != "" {
		if regExpPt, err = regexp.Compile(regExp); err != nil {
			return error116.Errorf("regexp.Compile: %w", err)
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
func (lg *LogInstance) IsThisDebug() bool {
	if atomic.LoadUint32(&lg.isDebug) != 0 {
		return true
	}
	regExp := lg.getRegexp()
	if regExp == nil {
		return false
	}
	cloc := runt.NewCodeLocation(lg.stackFramesToSkip - isThisDebugDelta)
	return regExp.MatchString(cloc.FuncName)
}

// IsSilent if true it means that Info does not print
func (lg *LogInstance) IsSilent() (isSilent bool) {
	return atomic.LoadUint32(&lg.isSilence) != 0
}

func (lg *LogInstance) getRegexp() *regexp.Regexp {
	lg.infoLock.RLock()
	defer lg.infoLock.RUnlock()
	return lg.infoRegexp
}

func (lg *LogInstance) doLog(format string, a ...interface{}) {
	s := sprintf(format, a...)
	if atomic.LoadUint32(&lg.isDebug) != 0 {
		s = appendLocation(s, runt.NewCodeLocation(lg.stackFramesToSkip))
	}
	if err := lg.output(0, s); err != nil {
		panic(error116.Errorf("LogInstance output: %w", err))
	}
}

func appendLocation(s string, location *runt.CodeLocation) string {
	// insert code location before a possible ending newline
	sNewline := ""
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
		sNewline = "\n"
	}
	return fmt.Sprintf("%s %s%s", s, location.Short(), sNewline)
}
