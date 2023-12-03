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
	// [StatusTerminal.CopyLog] stops output for a writer
	CopyLogRemoveFile = true
	// [pterm.StatusTerminalFd] default file descriptor for output
	STDefaultFd = 0
)

// [pterm.StatusTerminalFd] no write function
var STDefaultWriter io.Writer

// ANSI escape sequences
const (
	EraseEndOfLine         = "\x1b[K"
	EraseEndOfDisplay      = "\x1b[J"
	MoveCursorToColumnZero = "\r"
	CursorUp               = "\x1b[A" // at top line does nothing
	NewLine                = "\n"     // on last line, causes the view to scroll
	Space                  = "\x20"
)

// the Write signature of an io.Writer
type WriterWrite func(p []byte) (n int, err error)

// StatusTerminal provides an updatable status area at the bottom
// of the terminal window with log lines flowing upwards above it
//   - outputs to a single file descriptor or supplied Write method
//   - typically the os.Stderr is used
//   - uses [ANSI escape codes], ie. the stream is expected to
//     support certain ioctl
//   - the stream shpould typically not be piped or written to file,
//     since status output involves large amount of control codes
//   - supports ncurses-like behavior without putting
//     the terminal in a special mode
//   - when in status mode, each Print and Write invocations must
//     output in discrete lines
//   - LogTimeStamp prepends compact and specific timestamping
//
// [ANSI escape codes]: https://en.wikipedia.org/wiki/ANSI_escape_code
type StatusTerminal struct {
	// file descriptor, typically os.Stderr
	Fd int
	// the Printf-style function used for output
	Print func(s string)
	Write WriterWrite
	// whether fd actually is a terminal that can display status
	//	- true: stderr has ansi capabilities
	//	- false: stderr is a non-terminal pipe. width cannot be determined
	isTermTerminal bool

	// configurable whether status texts should be output
	IsTerminal atomic.Bool
	width      atomic.Int64 // atomic

	// no more status should be output
	statusEnded atomic.Bool

	lock             sync.Mutex
	displayLineCount int                // behind lock: number of terminal lines occupied by the current status
	output           string             // behind lock: the current status
	copyLog          map[io.Writer]bool // behind lock: log-copy streams
}

// NewStatusTerminal returns a terminal representation for
// concurrent logging and status output to standard error
//   - stores self-referencing pointers
func NewStatusTerminal() (statusTerminal *StatusTerminal) {
	return NewStatusTerminalFd(nil, 0, nil)
}

// NewStatusTerminalFd returns a terminal representation for logging and status output
//   - Fd is file descriptor used for ioctl, ie. needs to be a terminal, default is standard error
//   - Writer is for output, default writer for Fd
//   - stores self-referencing pointers
func NewStatusTerminalFd(fieldp *StatusTerminal, fd int, writer io.Writer, copyLog ...io.Writer) (statusTerminal *StatusTerminal) {

	// get statusTerminal value
	if fieldp != nil {
		statusTerminal = fieldp
		*statusTerminal = StatusTerminal{}
	} else {
		statusTerminal = &StatusTerminal{}
	}

	// get file descriptor to use for output
	var stderrFd = int(os.Stderr.Fd())
	if fd != 0 {
		statusTerminal.Fd = fd
	} else {
		statusTerminal.Fd = stderrFd
	}

	// get Write and Print methods
	if writer != nil {
		statusTerminal.Write = writer.Write
		statusTerminal.Print = statusTerminal.printWrite
	} else {
		var stdoutFd = int(os.Stdout.Fd())
		switch statusTerminal.Fd {
		case stderrFd:
			statusTerminal.Print = statusTerminal.printStderr
		case stdoutFd:
			statusTerminal.Print = statusTerminal.printStdout
		default:
			statusTerminal.Write = os.NewFile(uintptr(statusTerminal.Fd), "fd").Write
			statusTerminal.Print = statusTerminal.printWrite
		}
	}

	// handle log-copy streams
	if len(copyLog) > 0 {
		statusTerminal.copyLog = make(map[io.Writer]bool)
		statusTerminal.copyLog[copyLog[0]] = true
	}

	// IsTerminal
	//	- isTermTerminal is if the stream actually is a terminal
	//	- IsTerminal is whether the stream should be treated as a terminal
	if statusTerminal.isTermTerminal = term.IsTerminal(statusTerminal.Fd); statusTerminal.isTermTerminal {
		statusTerminal.IsTerminal.Store(true)
	}

	return
}

// Status updates a status area at the bottom of the display
//   - For non-ansi-terminal stderr, Status does nothing.
func (s *StatusTerminal) Status(statusLines string) {
	if !s.IsTerminal.Load() || s.statusEnded.Load() {
		return // no status if not terminal or EndStatus
	}
	width := s.Width()
	if width == 0 {
		return // zero window width return
	}

	// split s into lines
	var lines = strings.Split(statusLines, NewLine) // empty string has slice length 1, empty line

	// StatusDebug is used to troubleshoot status line counting
	//	- without invoking NewStatusDebug, d does nothing
	//	- d.n is true if d is active
	//	- d appends a text to the end of the last status line:
	//	- “01n02e03w004L05N06”
	//	- to activate d, use option “-verbose StatusTerminal..Status”
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

	s.lock.Lock()
	defer s.lock.Unlock()

	s.Print(s.clearStatus() + output + EraseEndOfDisplay)

	// save display status
	s.output = output
	s.displayLineCount = displayLineCount
}

// LogTimeStamp outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
//   - for two or more arguments, Printf formatting is used.
//   - Single argument is not interpreted.
//   - The string is preceed by a timestamp and space: "060102 15:04:05-08 "
//   - For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
func (s *StatusTerminal) LogTimeStamp(format string, a ...any) {
	s.Log(parl.ShortSpace() + Space + parl.Sprintf(format, a...))
}

// Log outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
// for two or more arguments, Printf formatting is used.
// Single argument is not interpreted.
// For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
func (s *StatusTerminal) Log(format string, a ...any) {

	// if not a terminal, regular logging
	if !s.IsTerminal.Load() {
		var logLines = parl.Sprintf(format, a...)
		s.Print(logLines) // parl.Log is thread-safe
		for writer := range s.copyLog {
			s.write(logLines, writer.Write)
		}
		return
	}

	// get log string that ends with newline
	var logLines string
	if len(a) > 0 {
		logLines = parl.Sprintf(format, a...)
	} else {
		logLines = format
	}
	if len(logLines) > 0 && logLines[len(logLines)-1:] != NewLine {
		logLines += NewLine
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.statusEnded.Load() {
		s.Print(s.clearStatus() + logLines + s.restoreStatus())
	} else {
		s.Print(logLines)
	}
	for writer := range s.copyLog {
		s.write(logLines, writer.Write)
	}
}

// SetTerminal overrides status regardless of whether a terminal is used
//   - isTerminal overrides the detection of if ANSI sequences are supported
//   - width is width to use if width cannot be read from the stream
//   - if width is -1, reset back to default
func (s *StatusTerminal) SetTerminal(isTerminal bool, width int) {
	if width == -1 {
		s.IsTerminal.Store(s.isTermTerminal)
		return
	}
	if isTerminal {
		if width < 1 {
			width = 1
		}
		s.width.Store(int64(width))
	}
	s.IsTerminal.Store(isTerminal)
}

// Width returns the current column width of the window
//   - if not a terminal, the stored width value
func (s *StatusTerminal) Width() (width int) {
	if !s.isTermTerminal {
		width = int(s.width.Load())
		return
	}
	var err error
	if width, _, err = term.GetSize(s.Fd); err != nil {
		panic(perrors.ErrorfPF("term.GetSize %w", err))
	}
	return
}

// CopyLog adds writers that receives copies of non-status logging
//   - remove true stops output for a writer
func (s *StatusTerminal) CopyLog(writer io.Writer, remove ...bool) {
	if writer == nil {
		panic(perrors.NewPF("writer cannot be nil"))
	}
	var delete0 bool
	if len(remove) > 0 {
		delete0 = remove[0]
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if !delete0 {
		if s.copyLog == nil {
			s.copyLog = map[io.Writer]bool{}
		}
		s.copyLog[writer] = true
		return
	}

	if s.copyLog != nil {
		delete(s.copyLog, writer)
	}
}

// EndStatus stops status output, typically on app exit
func (s *StatusTerminal) EndStatus() {
	if s.statusEnded.Load() {
		return //already end
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.statusEnded.CompareAndSwap(false, true) {
		return // did not win shutdown return
	}
	s.output = ""
	s.Print(NewLine)
}

// clearStatus returns ANSI codes to clear the status area if any
func (s *StatusTerminal) clearStatus() (clearStatusSequence string) {
	if len(s.output) > 0 {
		clearStatusSequence = MoveCursorToColumnZero + strings.Repeat(CursorUp, s.displayLineCount) + EraseEndOfDisplay
	}
	return
}

// restoreStatus returns the ANSI code to restore status if any
func (s *StatusTerminal) restoreStatus() (restoreStatusSequence string) {
	if len(s.output) > 0 {
		restoreStatusSequence = s.output + EraseEndOfDisplay
	}
	return
}

// printWrite outputs a string to the io.Write writer
//   - exposed as [StatusTerminal.Print]
func (s *StatusTerminal) printWrite(str string) { s.write(str, s.Write) }

// write writes a string as bytes to write-function writer
func (s *StatusTerminal) write(str string, writer WriterWrite) {
	var byts = []byte(str)
	var totalBytes = len(byts)
	var bytesWritten int

	// write until all written
	var n int
	var err error
	for bytesWritten < totalBytes {
		if n, err = writer(byts[bytesWritten:]); perrors.IsPF(&err, "provided Writer.Write %w", err) {
			panic(err)
		}
		bytesWritten += n
	}
}

// printStdout outputs to standard out via Parl, adapting the signature
func (s *StatusTerminal) printStdout(str string) { parl.Outw(str) }

// printStdout outputs to standard error via Parl, adapting the signature
func (s *StatusTerminal) printStderr(str string) { parl.Logw(str) }
