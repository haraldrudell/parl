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
	"github.com/haraldrudell/parl/pos"
	"golang.org/x/term"
)

const (
	// [pterm.StatusTerminalFd] default file descriptor for output
	STDefaultFd int = 0
	// [StatusTerminal.SetTerminal] width for reset to default
	DisableSetTerminal int = -1
	// [StatusTerminal.SetTerminal] default isTerminal value
	NoIsTerminal IsTerminal = false
	// [StatusTerminal.SetTerminal] activate isTerminal override
	IsTerminalYes IsTerminal = true
	// [StatusTerminal.CopyLog] remove writer
	CopyLogRemove CopyLogRemover = true
)

// determines status output override for [StatusTerminal.SetTerminal]
//   - [IsTerminalYes] [NoIsTerminal]
type IsTerminal bool

// argument type for [StatusTerminal.CopyLog]
//   - [CopyLogRemove]
type CopyLogRemover bool

// [pterm.StatusTerminalFd] no write function
var STDefaultWriter io.Writer

// [pterm.StatusTerminalFd] no printf function
var NoPrintf parl.PrintfFunc

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
	// file descriptor for status output, typically [os.Stderr]
	//	- output is to Write or Print, those possibly created based on fd
	//	- fd is required to invoke [term.IsTerminal] and [term.GetSize]
	//	- after new-function, fd is never STDefaultFd
	fd int
	// the Printf-style function used for status output
	//	- logPrinter is error-free Write taking a string
	//	- empty string is noop
	//	- either printStderr printStdout or printWrite
	//	- must hold lock to ensure clear-stdout output-restore sequence
	logPrinter Printer
	// output to standard output
	stdoutPrinter Printer
	// whether fd actually is a terminal that can display status
	//	- the value from [term.IsTerminal]
	//	- value used is isTerminal that may be overriden
	//	- true: StatusTerminal.fd has ansi capabilities
	//	- false: StatusTerminal.fd is a non-terminal pipe.
	//		Width cannot be determined
	isTermTerminal bool

	// configurable value whether status texts should be output
	isTerminal atomic.Bool
	// width for non-pterm.IsTerminal. Returned by Width method
	width parl.Atomic64[int]

	// no more status should be output
	statusEnded atomic.Bool

	// make certain values thread-safe
	//	- atomizes multiple [StatusTerminal.print] invocations
	lock sync.Mutex
	// behind lock: number of terminal lines occupied by the current status
	displayLineCount int
	// behind lock: the current status
	output string
	// behind lock: log-copy streams
	copyLog map[io.Writer]bool
}

// NewStatusTerminal returns a terminal representation for
// concurrent logging and status output to standard error
//   - output is to stderr and stdout
//   - status enabled if stderr is terminal
//   - stores self-referencing pointers
//   - [StatusTerminal.Status] set status string
//   - [StatusTerminal.Log] stderr output no timestamp
//   - [StatusTerminal.LogStdout] stdout output no timestamp
//   - [StatusTerminal.LogTimeStamp] stderr output with timestamp
func NewStatusTerminal() (statusTerminal *StatusTerminal) {
	return NewStatusTerminalFd(
		nil,
		0,
		STDefaultWriter, STDefaultWriter,
		NoPrintf, NoPrintf,
	)
}

// NewStatusTerminalFd returns a terminal representation for logging and status output
//   - fieldp: pointer to optional pre-allocated field
//   - fd: file descriptor used for ioctl, ie. needs to be a terminal
//   - — status output is written to fd if logWriter not present
//   - — fd is used for detecting ANSI capabilities and display width
//   - fd STDefaultFd 0: use default standard error: os.Stderr.Fd()
//   - logWriter: used for status output
//   - logWriter nil: output to fd
//   - stdoutWriter: used for LogStdout output
//   - stdoutWriter nil: LogStdout outputs to standard output
//   - logPrintf, stdoutPrintf: set or NoPrintf
//   - copyLog optional: log output should be copied to this writer
//   - stores self-referencing pointers
//   - [StatusTerminal.Status] set status string
//   - [StatusTerminal.Log] stderr output no timestamp
//   - [StatusTerminal.LogStdout] stdout output no timestamp
//   - [StatusTerminal.LogTimeStamp] stderr output with timestamp
func NewStatusTerminalFd(
	fieldp *StatusTerminal,
	fd int,
	logWriter io.Writer,
	stdoutWriter io.Writer,
	logPrintf parl.PrintfFunc,
	stdoutPrintf parl.PrintfFunc,
	copyLog ...io.Writer,
) (statusTerminal *StatusTerminal) {

	// statusTerminal points to struct value
	if fieldp != nil {
		statusTerminal = fieldp
		*statusTerminal = StatusTerminal{}
	} else {
		statusTerminal = &StatusTerminal{}
	}

	var stderrFd = int(os.Stderr.Fd())
	if stderrFd == 0 {
		panic(perrors.NewPF("stderr fd zero"))
	}
	var stdoutFd = int(os.Stdout.Fd())

	// statusTerminal.fd is file descriptor for status output
	//	- provided non-zero fd or os.Stderr.Fd
	if fd != STDefaultFd /*0*/ {
		statusTerminal.fd = fd
	} else {
		statusTerminal.fd = stderrFd
	}

	// set statusTerminal.logPrinter
	if logWriter != nil {
		statusTerminal.logPrinter.SetWriter(logWriter)
	} else if logPrintf != nil {
		statusTerminal.logPrinter.SetPrintFunc(logPrintf)
	} else {
		switch statusTerminal.fd {
		case stderrFd:
			statusTerminal.logPrinter.SetPrintFunc(parl.Logw)
		case stdoutFd:
			statusTerminal.logPrinter.SetPrintFunc(parl.Outw)
		default:
			statusTerminal.logPrinter.SetWriter(os.NewFile(uintptr(statusTerminal.fd), "fd"))
		}
	}

	// set statusTerminal.stdoutPrinter
	if stdoutWriter != nil {
		statusTerminal.stdoutPrinter.SetWriter(stdoutWriter)
	} else if stdoutPrintf != nil {
		statusTerminal.stdoutPrinter.SetPrintFunc(stdoutPrintf)
	} else {
		statusTerminal.stdoutPrinter.SetPrintFunc(parl.Outw)
	}

	// add possible copyLog to statusTerminal.copyLog map
	if len(copyLog) > 0 {
		statusTerminal.copyLog = make(map[io.Writer]bool)
		statusTerminal.copyLog[copyLog[0]] = true
	}

	// set statusTerminal.isTerminal
	//	- isTermTerminal is if the stream actually is a terminal
	//	- IsTerminal is whether the stream should be treated as a terminal
	if statusTerminal.isTermTerminal = term.IsTerminal(statusTerminal.fd); statusTerminal.isTermTerminal {
		statusTerminal.isTerminal.Store(true)
	}

	return
}

// Status updates a status area at the bottom of the display
//   - For non-ansi-terminal stderr, Status does nothing
//   - delegates to logPrinter
//   - thread-safe
func (s *StatusTerminal) Status(statusLines string) {

	// check if status should be output
	if !s.isTerminal.Load() || s.statusEnded.Load() {
		return // no status if not terminal or EndStatus
	}

	// only output for non-zero window size
	var width = s.Width()
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
	var length = len(lines)
	var i = length // 1…
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
	var displayLineCount = 0
	var output = ""
	var lastIndex = len(lines) - 1
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

	s.logPrinter.Print(s.clearStatus() + output + EraseEndOfDisplay)

	// save display status
	s.output = output
	s.displayLineCount = displayLineCount
}

// LogTimeStamp outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
//   - two or more arguments: Printf formatting is used
//   - Single argument is not interpreted
//   - The string is preceed by a timestamp and space: "060102 15:04:05-08 "
//   - For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
//   - thread-safe
func (s *StatusTerminal) LogTimeStamp(format string, a ...any) {
	s.Log(parl.ShortSpace() + Space + parl.Sprintf(format, a...))
}

// Log outputs text ending with at least one newline while maintaining status information
// at bottom of screen.
//   - default to stderr
//   - two or more arguments: Printf formatting used
//   - single argument is not interpreted
//   - For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
//   - thread-safe
func (s *StatusTerminal) Log(format string, a ...any) { s.doLog(pos.Stderr, format, a...) }

// LogStdout outputs to specific logger, ie. stdout
//   - default to stdout
//   - two or more arguments: Printf formatting used
//   - single argument is not interpreted
//   - thread-safe
func (s *StatusTerminal) LogStdout(format string, a ...any) { s.doLog(pos.Stdout, format, a...) }

// SetTerminal overrides status regardless of whether a terminal is used
//   - isTerminal: overrides the detection of whether ANSI sequences are supported
//   - isTerminal IsTerminalYes true: enable override
//   - isTerminal NoIsTerminal false: disable override or noop
//   - width: width to use if fd is not ANSI terminal
//   - width DisableSetTerminal -1: disable SetTerminal override
//   - thread-safe
func (s *StatusTerminal) SetTerminal(isTerminal IsTerminal, width int) (isAnsi bool) {
	isAnsi = s.isTermTerminal

	// handle reset case
	if width == -1 {
		// check whether write required
		if s.isTerminal.Load() == s.isTermTerminal {
			return
		}
		s.isTerminal.Store(s.isTermTerminal)
		return
	}

	// set fake width
	if isTerminal {
		if width < 1 {
			width = 1
		}
		s.width.Store(width)
	}

	// set SetTerminal override
	s.isTerminal.Store(isTerminal == IsTerminalYes)

	return
}

// Width returns the current column width of the window
//   - if SetTerminal override active, its provided width value
//   - thread-safe
func (s *StatusTerminal) Width() (width int) {

	// if not actually ANSI terminal, use fake width
	if !s.isTermTerminal {
		width = int(s.width.Load())
		return
	}

	// get ANSI width
	var err error
	if width, _, err = term.GetSize(s.fd); err != nil {
		panic(perrors.ErrorfPF("term.GetSize %w", err))
	}

	return
}

// CopyLog adds writers that receives copies of non-status logging
//   - writer: an [io.Writer] receiving log output
//   - remove: present and CopyLogRemove true: stop output to a previous writer
//   - thread-safe
func (s *StatusTerminal) CopyLog(writer io.Writer, remove ...CopyLogRemover) {

	// ensure writer present
	if writer == nil {
		panic(perrors.NewPF("writer cannot be nil"))
	}

	// deteremine if delete case
	var isDeleteWriter CopyLogRemover
	if len(remove) > 0 {
		isDeleteWriter = remove[0]
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	// handle add case
	if isDeleteWriter != CopyLogRemove {
		if s.copyLog == nil {
			s.copyLog = map[io.Writer]bool{}
		}
		s.copyLog[writer] = true
		return
	}

	// delete case
	if s.copyLog != nil {
		delete(s.copyLog, writer)
	}
}

// EndStatus stops status output, typically on app exit
//   - idempotent
//   - thread-safe
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
	s.logPrinter.Print(NewLine)
}

// NeedsEndStatus is true if a ne- function has been invoked and
// EndStatus has not been invoked
//   - thread-safe
func (s *StatusTerminal) NeedsEndStatus() (needsEndStatus bool) {
	return s.fd != STDefaultFd && !s.statusEnded.Load()
}

// doLog implements: LogTimestamp Log LogStdout
//   - print formatted to the proper output stream
//     and output to copyLog writers
//   - ensures terminating newline
//   - the Status method does not arrive here,
//     Status outputs directly to print
func (s *StatusTerminal) doLog(standardStream pos.StandardStream, format string, a ...any) {

	// printf to single string, ensure ending with newline
	var logLinesNewline = parl.Sprintf(format, a...)
	if len(logLinesNewline) == 0 || logLinesNewline[len(logLinesNewline)-1:] != NewLine {
		logLinesNewline += NewLine
	}
	// text is not empty

	// delegate to print, printStdout or doStatus
	if !s.isTerminal.Load() || s.statusEnded.Load() {
		s.doPrint(standardStream, logLinesNewline)
	} else {
		// output to status terminal
		s.doStatus(standardStream, logLinesNewline)
	}

	// copy to log-copy writers
	for writer := range s.copyLog {
		s.logFileWrite(logLinesNewline, writer)
	}
}

// doPrint outputs the non-status case
func (s *StatusTerminal) doPrint(standardStream pos.StandardStream, logLines string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// if not a terminal, regular logging
	if standardStream == pos.Stderr {
		s.logPrinter.Print(logLines)
	} else {
		s.stdoutPrinter.Print(logLines)
	}
}

// doStatus implements LogTimestamp Log LogStdout display output
// while status output is active
func (s *StatusTerminal) doStatus(standardStream pos.StandardStream, logLines string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// simple stderr case
	if standardStream == pos.Stderr {
		s.logPrinter.Print(s.clearStatus() + logLines + s.restoreStatus())
		return
	}

	// stdout case
	s.logPrinter.Print(s.clearStatus())
	s.stdoutPrinter.Print(logLines)
	s.logPrinter.Print(s.restoreStatus())
}

// clearStatus returns ANSI codes to clear the status area if any
//   - may be empty string
func (s *StatusTerminal) clearStatus() (clearStatusSequence string) {
	if len(s.output) > 0 {
		clearStatusSequence = MoveCursorToColumnZero + strings.Repeat(CursorUp, s.displayLineCount) + EraseEndOfDisplay
	}
	return
}

// restoreStatus returns the ANSI code to restore status if any
// - may be empty string
func (s *StatusTerminal) restoreStatus() (restoreStatusSequence string) {
	if len(s.output) > 0 {
		restoreStatusSequence = s.output + EraseEndOfDisplay
	}
	return
}

// write writes a string as bytes to [io.Writer.Write] function
//   - panic on Write error
func (s *StatusTerminal) logFileWrite(str string, writer io.Writer) {
	var byts = []byte(str)
	var totalBytes = len(byts)
	var bytesWritten int

	// write until all written
	var n int
	var err error
	for bytesWritten < totalBytes {
		if n, err = writer.Write(byts[bytesWritten:]); perrors.IsPF(&err, "provided Writer.Write %w", err) {
			panic(err)
		}
		bytesWritten += n
	}
}
