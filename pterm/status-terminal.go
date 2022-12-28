/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"os"
	"strings"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/term"
)

const (
	eraseEndOfLine         = "\x1b[K"
	eraseEndOfDisplay      = "\x1b[J"
	moveCursorToColumnZero = "\r"
	cursorUp               = "\x1b[A"
	newLine                = "\n"
	space                  = "\x20"
)

// StatusTerminal is a terminal supporting both log output and
// status information using [ANSI escape codes].
// StatusTerminal supports ncurses-like behaviors but does not
// put the terminal in a special mode.
//
// [ANSI escape codes]: https://en.wikipedia.org/wiki/ANSI_escape_code
type StatusTerminal struct {
	Fd         int  // file descripor, typically os.Stderr
	IsTerminal bool // true: stderr has ansi capabilities. false: stderr is a non-terminal pipe

	lock             sync.Mutex
	displayLineCount int    // behind lock: numer of terminal lines occupied by the current status
	output           string // behind lock: the current status
}

// NewStatusTerminal returns a terminal represenation for loggins and status output to stderr
func NewStatusTerminal() (statusTerminal *StatusTerminal) {
	fd := int(os.Stderr.Fd())
	return &StatusTerminal{
		Fd:         fd,
		IsTerminal: term.IsTerminal(fd),
	}
}

// Status updates a status are at the bottom of the display.
// For non-ansi-terminal stderr, Status does nothing.
func (st *StatusTerminal) Status(s string) {
	if !st.IsTerminal {
		return // no status if not terminal
	}
	width, _, err := term.GetSize(st.Fd)
	if err != nil {
		panic(perrors.ErrorfPF("term.GetSize %w", err))
	}

	// split s into lines
	lines := strings.Split(s, newLine) // empty string has slice length 1, empty line

	// remove trailing blank lines
	length := len(lines)
	i := length // 1…
	for i > 0 {
		if len(lines[i-1]) != 0 {
			break // a non-empty line
		}
		i--
	}
	// i now a valid index 0…
	if i != length {
		lines = lines[:i] // lines has no trailing blank lines, length may be 0
	}

	// get displayLineCount and output
	displayLineCount := 0
	output := ""
	lastIndex := len(lines) - 1
	for i, line := range lines {
		printablesLine := TrimANSIEscapes(line)
		displayLineCount += len(printablesLine)/width + 1
		output += line
		if i < lastIndex {
			isAtDisplayWidth := len(printablesLine) > 0 && len(printablesLine)%width == 0
			if !isAtDisplayWidth {
				output += eraseEndOfLine + newLine
			}
		} else {
			displayLineCount-- // last line has no newline
		}
	}
	output += space // places cursor one step to the right of final output character

	st.lock.Lock()
	defer st.lock.Unlock()

	parl.Logw(st.clearStatus() + output + eraseEndOfDisplay)

	// save display status
	st.output = output
	st.displayLineCount = displayLineCount
}

func (st *StatusTerminal) clearStatus() (s string) {
	if len(st.output) > 0 {
		s = moveCursorToColumnZero + strings.Repeat(cursorUp, st.displayLineCount) + eraseEndOfDisplay
	}
	return
}

func (st *StatusTerminal) restoreStatus() (s string) {
	if len(st.output) > 0 {
		s = st.output + eraseEndOfDisplay
	}
	return
}

// Log outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
// for two or more arguments, Printf formatting is used.
// Single argument is not interpreted.
// The string is preceed by a timestamp and space: "060102 15:04:05-08 "
// For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
func (st *StatusTerminal) LogTimeStamp(format string, a ...interface{}) {
	st.Log(parl.ShortSpace() + space + parl.Sprintf(format, a...))
}

// Log outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
// for two or more arguments, Printf formatting is used.
// Single argument is not interpreted.
// For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
func (st *StatusTerminal) Log(format string, a ...interface{}) {
	if !st.IsTerminal {
		parl.Log(format, a...) // parl.Log is thread-safe
		return
	}

	// get log string that ends with newline
	var s string
	if len(a) > 0 {
		s = parl.Sprintf(format, a...)
	} else {
		s = format
	}
	if len(s) > 0 && s[len(s)-1:] != newLine {
		s += newLine
	}

	st.lock.Lock()
	defer st.lock.Unlock()

	parl.Logw(st.clearStatus() + s + st.restoreStatus())
}
