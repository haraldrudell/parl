/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"

	"github.com/haraldrudell/parl/ptermx"
)

// StatusTerminal is interface for [pterm.StatusTerminal]
type StatusTerminal interface {
	// Status updates a status area at the bottom of the display
	//   - For non-ansi-terminal stderr, Status does nothing
	//   - delegates to logPrinter
	//   - thread-safe
	Status(statusLines string)
	// LogTimeStamp outputs text ending with at least one newline while maintaining status information
	// at bottom of screen.
	//   - two or more arguments: Printf formatting is used
	//   - Single argument is not interpreted
	//   - The string is preceed by a timestamp and space: "060102 15:04:05-08 "
	//   - For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
	//   - thread-safe
	LogTimeStamp(format string, a ...any)
	// Log outputs text ending with at least one newline while maintaining status information
	// at bottom of screen.
	//   - default to stderr
	//   - two or more arguments: Printf formatting used
	//   - single argument is not interpreted
	//   - For non-ansi-terminal stderr, LogTimeStamp simply prints lines of text.
	//   - thread-safe
	Log(format string, a ...any)
	// LogStdout outputs to specific logger, ie. stdout
	//   - default to stdout
	//   - two or more arguments: Printf formatting used
	//   - single argument is not interpreted
	//   - thread-safe
	LogStdout(format string, a ...any)
	// SetTerminal overrides status regardless of whether a terminal is used
	//   - isTerminal: overrides the detection of whether ANSI sequences are supported
	//   - isTerminal IsTerminalYes true: enable override
	//   - isTerminal NoIsTerminal false: disable override or noop
	//   - width: width to use if fd is not ANSI terminal
	//   - width DisableSetTerminal -1: disable SetTerminal override
	//   - thread-safe
	SetTerminal(isTerminal IsTerminal, width int) (isAnsi bool)
	// Width returns the current column width of the window
	//   - if SetTerminal override active, its provided width value
	//   - thread-safe
	Width() (width int)
	// CopyLog adds writers that receives copies of non-status logging
	//   - writer: an [io.Writer] receiving log output
	//   - remove: present and CopyLogRemove true: stop output to a previous writer
	//   - thread-safe
	CopyLog(writer io.Writer, remove ...CopyLogRemover)
	// EndStatus stops status output, typically on app exit
	//   - idempotent
	//   - thread-safe
	EndStatus()
	// NeedsEndStatus is true if a ne- function has been invoked and
	// EndStatus has not been invoked
	//   - thread-safe
	NeedsEndStatus() (needsEndStatus bool)
	// State returns state for debug purposes
	State() (s2 ptermx.StatusTerminalState)
}

const (
	// [StatusTerminal.CopyLog] remove writer
	CopyLogRemove = ptermx.CopyLogRemove
)

// argument type for [StatusTerminal.CopyLog]
// - [CopyLogRemove]
type CopyLogRemover = ptermx.CopyLogRemover

const (
	// [StatusTerminal.SetTerminal] default isTerminal value
	NoIsTerminal = ptermx.NoIsTerminal
	// [StatusTerminal.SetTerminal] activate isTerminal override
	IsTerminalYes = ptermx.IsTerminalYes
)

// determines status output override for [StatusTerminal.SetTerminal]
//   - [IsTerminalYes] [NoIsTerminal]
type IsTerminal = ptermx.IsTerminal
