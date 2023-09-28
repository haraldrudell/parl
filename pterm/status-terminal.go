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
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/term"
)

const (
	EraseEndOfLine         = "\x1b[K"
	EraseEndOfDisplay      = "\x1b[J"
	MoveCursorToColumnZero = "\r"
	CursorUp               = "\x1b[A" // at top line does nothing
	NewLine                = "\n"     // on last line, causes the view to scroll
	Space                  = "\x20"
)

type WriterWrite func(p []byte) (n int, err error)

// StatusTerminal is a terminal supporting both log output and
// status information using [ANSI escape codes].
// StatusTerminal supports ncurses-like behaviors but does not
// put the terminal in a special mode.
//
// [ANSI escape codes]: https://en.wikipedia.org/wiki/ANSI_escape_code
type StatusTerminal struct {
	Fd    int            // file descriptor, typically os.Stderr
	Print func(s string) // the Printf-style function used for output
	Write WriterWrite
	// whether fd actually is a terminal
	// true: stderr has ansi capabilities. false: stderr is a non-terminal pipe
	isTermTerminal bool

	IsTerminal parl.AtomicBool // configurable whether status texts should be output
	width      int64           // atomic

	statusEnded parl.AtomicBool // no more status should be output

	lock             sync.Mutex
	displayLineCount int    // behind lock: number of terminal lines occupied by the current status
	output           string // behind lock: the current status
	copyLog          map[io.Writer]bool
}

// NewStatusTerminal returns a terminal representation for logging and status output to stderr.
func NewStatusTerminal() (statusTerminal *StatusTerminal) {
	return new(0, nil)
}

// NewStatusTerminalFd returns a terminal representation for logging and status output.
//   - Fd is file descriptor used for ioctl, ie. needs to be a terminal, default stderr
//   - Writer is for output, default writer for Fd
func NewStatusTerminalFd(fd int, writer io.Writer) (statusTerminal *StatusTerminal) {
	return new(fd, nil)
}

func new(fd int, writer io.Writer) (statusTerminal *StatusTerminal) {

	// fd
	stdoutFd := int(os.Stderr.Fd())
	stderrFd := int(os.Stderr.Fd())
	if fd == 0 {
		fd = stderrFd
	}

	s := StatusTerminal{Fd: fd}

	if writer != nil {
		s.Write = writer.Write
		s.Print = s.printWrite
	} else {
		if fd == stderrFd {
			s.Print = s.printStderr
		} else if fd == stdoutFd {
			s.Print = s.printStdout
		} else {
			s.Write = os.NewFile(uintptr(fd), "fd").Write
			s.Print = s.printWrite
		}
	}

	// IsTerminal
	if s.isTermTerminal = term.IsTerminal(fd); s.isTermTerminal {
		s.IsTerminal.Set()
	}

	return &s
}

// Status updates a status are at the bottom of the display.
// For non-ansi-terminal stderr, Status does nothing.
func (st *StatusTerminal) Status(s string) {
	if !st.IsTerminal.IsTrue() || st.statusEnded.IsTrue() {
		return // no status if not terminal or EndStatus
	}
	width := st.Width()
	if width == 0 {
		return // zero window width return
	}

	// split s into lines
	var lines = strings.Split(s, NewLine) // empty string has slice length 1, empty line

	// StatusDebug is used to troubleshoot status line counting
	//	- without invoking NewStatusDebug, d does nothing
	//	- d.n is true if d is active
	//	- d appends a text to the end of the last status line:
	//	- “01n02e03w004L05N06”
	//	- to activate d, use option “-verbose StatusTerminal.Status”
	var d = StatusDebug{}
	if parl.IsThisDebug() {
		d = *NewStatusDebug(lines, width)
		// insert a pre-release of debug data to have fixed status length
		lines[len(lines)-1] += d.DebugText()
	}

	// remove trailing blank lines
	length := len(lines)
	i := length // 1…
	for i > 0 {
		if len(lines[i-1]) != 0 {
			break // a non-empty line
		}
		i--
	}
	d.metaEmptyLineCount = len(lines) - i
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
			d.metaLongLines += (length - 1) / width
		}
		output += line

		if i < lastIndex {
			if !cursorAtEndOfLine {
				output += EraseEndOfLine
			}
			displayLineCount++ // count the newline
			d.metaCountedNewlines++
			output += NewLine
		} else {
			if !cursorAtEndOfLine {
				output += Space // places cursor one step to the right of final output character
			}
		}
	}

	// update meta string
	if d.n {
		output = d.UpdateOutput(output, displayLineCount)
	}

	st.lock.Lock()
	defer st.lock.Unlock()

	st.Print(st.clearStatus() + output + EraseEndOfDisplay)

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
	st.Log(parl.ShortSpace() + Space + parl.Sprintf(format, a...))
}

// Log outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
// for two or more arguments, Printf formatting is used.
// Single argument is not interpreted.
// For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
func (st *StatusTerminal) Log(format string, a ...any) {
	if !st.IsTerminal.IsTrue() {
		var s = parl.Sprintf(format, a...)
		st.Print(s) // parl.Log is thread-safe
		for writer := range st.copyLog {
			st.write(s, writer.Write)
		}
		return
	}

	// get log string that ends with newline
	var s string
	if len(a) > 0 {
		s = parl.Sprintf(format, a...)
	} else {
		s = format
	}
	if len(s) > 0 && s[len(s)-1:] != NewLine {
		s += NewLine
	}

	st.lock.Lock()
	defer st.lock.Unlock()

	if st.statusEnded.IsFalse() {
		st.Print(st.clearStatus() + s + st.restoreStatus())
	} else {
		st.Print(s)
	}
	for writer := range st.copyLog {
		st.write(s, writer.Write)
	}
}

func (st *StatusTerminal) SetTerminal(isTerminal bool, width int) {
	if isTerminal {
		if width < 1 {
			width = 1
		}
		atomic.StoreInt64(&st.width, int64(width))
		st.IsTerminal.Set()
	} else {
		st.IsTerminal.Clear()
	}
}

func (st *StatusTerminal) Width() (width int) {
	if !st.isTermTerminal {
		width = int(atomic.LoadInt64(&st.width))
		return
	}
	var err error
	if width, _, err = term.GetSize(st.Fd); err != nil {
		panic(perrors.ErrorfPF("term.GetSize %w", err))
	}
	return
}

func (st *StatusTerminal) CopyLog(writer io.Writer, remove ...bool) {
	if writer == nil {
		panic(perrors.NewPF("writer cannot be nil"))
	}
	var delete0 bool
	if len(remove) > 0 {
		delete0 = remove[0]
	}

	st.lock.Lock()
	defer st.lock.Unlock()

	if !delete0 {
		if st.copyLog == nil {
			st.copyLog = map[io.Writer]bool{}
		}
		st.copyLog[writer] = true
		return
	}

	if st.copyLog != nil {
		delete(st.copyLog, writer)
	}
}

func (st *StatusTerminal) EndStatus() {
	if st.statusEnded.IsTrue() {
		return //already end
	}
	st.lock.Lock()
	defer st.lock.Unlock()

	if !st.statusEnded.Set() {
		return // did not win shutdown return
	}

	st.output = ""
	st.Print(NewLine)
}

func (st *StatusTerminal) clearStatus() (s string) {
	if len(st.output) > 0 {
		s = MoveCursorToColumnZero + strings.Repeat(CursorUp, st.displayLineCount) + EraseEndOfDisplay
	}
	return
}

func (st *StatusTerminal) restoreStatus() (s string) {
	if len(st.output) > 0 {
		s = st.output + EraseEndOfDisplay
	}
	return
}

func (st *StatusTerminal) printWrite(s string) {
	st.write(s, st.Write)
}

func (st *StatusTerminal) write(s string, writer WriterWrite) {
	byts := []byte(s)
	length := len(byts)
	var n0 int

	var n int
	var err error
	for n0 < length {
		if n, err = writer(byts[n0:]); perrors.IsPF(&err, "provided Writer.Write %w", err) {
			panic(err)
		}
		n0 += n
	}
}

func (st *StatusTerminal) printStdout(s string) {
	parl.Outw(s)
}

func (st *StatusTerminal) printStderr(s string) {
	parl.Logw(s)
}
