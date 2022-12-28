/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"fmt"
	"os"
	"strings"

	"github.com/haraldrudell/parl/plogger"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const pruntimeDframes = 1

// stderrLogLogger is a shared log.Logger instance for stderr.
//   - using this for output ensures thread-safety with plog and parl packages
var stderrLogLogger = plogger.GetLog(os.Stderr)

// messageSprintf is like fmt.Sprintf with thousands separator
var messageSprintf = message.NewPrinter(language.English).Sprintf

// sprintf is like fmt.Sprintf but does not interpret format if a is empty, and has thousands separator
func sprintf(format string, a ...interface{}) (s string) {
	if len(a) > 0 {
		s = messageSprintf(format, a...)
	} else {
		s = format
	}
	return
}

// D prints to stderr with code location. Thread-safe.
//   - pruntime.D is like parl.D but can be used by packages imported by parl
//   - pruntime imports from only one parl package: plogger
//   - D is meant for temporary output intended to be removed before check-in
func D(format string, a ...interface{}) {
	if err := stderrLogLogger.Output(0,
		appendLocation(sprintf(format, a...), NewCodeLocation(pruntimeDframes))); err != nil {
		panic(fmt.Errorf("pruntime.D log.Logger.Output error: %w", err))
	}
}

// appendLocation appends code location at end of string and handles terminating newline
func appendLocation(s string, location *CodeLocation) string {
	// insert code location before a possible ending newline
	sNewline := ""
	if strings.HasSuffix(s, "\n") {
		s = s[:len(s)-1]
		sNewline = "\n"
	}
	return fmt.Sprintf("%s %s%s", s, location.Short(), sNewline)
}
