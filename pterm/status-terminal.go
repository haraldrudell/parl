/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"io"
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
	Fd         int            // file descriptor, typically os.Stderr
	Print      func(s string) // the Printf-style function used for output
	Write      func(p []byte) (n int, err error)
	IsTerminal bool // true: stderr has ansi capabilities. false: stderr is a non-terminal pipe

	lock             sync.Mutex
	displayLineCount int    // behind lock: numer of terminal lines occupied by the current status
	output           string // behind lock: the current status
}

// NewStatusTerminal returns a terminal represenation for loggins and status output to stderr
func NewStatusTerminal() (statusTerminal *StatusTerminal) {
	return new(0, nil, nil)
}

// NewStatusTerminalFd returns a terminal represenation for loggins and status output to stderr
func NewStatusTerminalFd(fd int, print func(s string)) (statusTerminal *StatusTerminal) {
	return new(fd, print, nil)
}

// NewStatusTerminalWriter returns a terminal represenation for loggins and status output to stderr
func NewStatusTerminalWriter(fd int, writer io.Writer) (statusTerminal *StatusTerminal) {
	return new(fd, nil, writer)
}

func new(fd int, print func(s string), writer io.Writer) (statusTerminal *StatusTerminal) {
	s := StatusTerminal{
		Fd:    fd,
		Print: print,
	}

	// fd
	stdoutFd := int(os.Stderr.Fd())
	stderrFd := int(os.Stderr.Fd())
	if fd == 0 {
		s.Fd = stderrFd
	}

	// print and Writer
	if print == nil {
		if writer != nil {
			s.Write = writer.Write
			s.Print = s.printWrite
		} else if s.Fd == stdoutFd {
			s.Print = s.printStdout
		} else if s.Fd == stderrFd {
			s.Print = s.printStderr
		} else {
			panic(perrors.NewPF("print and writer both nil, fd not os.Stdout or os.Stderr: please provide print or io.Writer"))
		}
	}

	// IsTerminal
	s.IsTerminal = term.IsTerminal(s.Fd)

	return &s
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
	if width == 0 {
		return // zero window width return
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
	// lines length is now i: 0…
	// lines ends with a non-empty line or has length zero

	// get displayLineCount and output
	displayLineCount := 0
	output := ""
	lastIndex := len(lines) - 1
	for i, line := range lines {
		printablesLine := TrimANSIEscapes(line)
		length := len([]rune(printablesLine)) // length in unicode code points
		var cursorAtEndOfLine bool
		// if length exactly matches width, cursor is still on the same line
		// therefore subtract 1 character from length
		if length > 0 {
			displayLineCount += (length - 1) / width
			cursorAtEndOfLine = length%width == 0
		}
		output += line

		if i < lastIndex {
			if !cursorAtEndOfLine {
				output += eraseEndOfLine
			}
			displayLineCount++ // count the newline
			output += newLine
		} else {
			if !cursorAtEndOfLine {
				output += space // places cursor one step to the right of final output character
			}
		}
	}

	st.lock.Lock()
	defer st.lock.Unlock()

	st.Print(st.clearStatus() + output + eraseEndOfDisplay)

	// save display status
	st.output = output
	st.displayLineCount = displayLineCount
}

// Log outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
// for two or more arguments, Printf formatting is used.
// Single argument is not interpreted.
// The string is preceed by a timestamp and space: "060102 15:04:05-08 "
// For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
func (st *StatusTerminal) LogTimeStamp(format string, a ...any) {
	st.Log(parl.ShortSpace() + space + parl.Sprintf(format, a...))
}

// Log outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
// for two or more arguments, Printf formatting is used.
// Single argument is not interpreted.
// For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
func (st *StatusTerminal) Log(format string, a ...any) {
	if !st.IsTerminal {
		st.Print(parl.Sprintf(format, a...)) // parl.Log is thread-safe
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

	st.Print(st.clearStatus() + s + st.restoreStatus())
}

func (st *StatusTerminal) Width() (width int) {
	var err error
	if width, _, err = term.GetSize(st.Fd); err != nil {
		panic(perrors.ErrorfPF("term.GetSize %w", err))
	}
	return
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

func (st *StatusTerminal) printWrite(s string) {
	var n int
	var err error
	if n, err = st.Write([]byte(s)); perrors.IsPF(&err, "provided Writer.Write %w", err) {
		panic(err)
	}
	_ = n
}

func (st *StatusTerminal) printStdout(s string) {
	parl.Outw(s)
}

func (st *StatusTerminal) printStderr(s string) {
	parl.Logw(s)
}
