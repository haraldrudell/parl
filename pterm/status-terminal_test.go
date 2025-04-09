/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"bytes"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/ptermx"
)

func TestStatusTerminal(t *testing.T) {
	const (
		widthToUse      = 256
		statusValue     = "#status"
		logValue        = "#log"
		stdoutValue     = "#stdout"
		expStdout       = stdoutValue + "\n"
		copyLogValue    = "#copyLog"
		expCopyLog      = copyLogValue + "\n"
		CR              = "\r"
		expStatusLines  = 5
		expStatus1      = logValue + "\n" + statusValue
		expStatus2      = "\x20" + logValue + "\n" + statusValue
		expStatus3      = statusValue
		expStatus4      = copyLogValue + "\n" + statusValue
		expStatusSuffix = "\n"
	)
	var (
		// stderrBuffer captures standard error output
		//	- [bytes.Buffer] is [io.Writer] to accessible byte-slice
		stderrBuffer bytes.Buffer
		// stdoutBuffer captures standard output
		stdoutBuffer bytes.Buffer
		// copyLogBuffer captures log-file output
		copyLogBuffer bytes.Buffer
		// noFieldp causes fresh allocation
		noFieldp *StatusTerminal
	)

	var (
		isAnsi         bool
		needsEndStatus bool
		width          int
		actStatus      string
		statusLines    []string
	)

	// CopyLog() EndStatus() Log() LogStdout() LogTimeStamp()
	// NeedsEndStatus() SetTerminal() Status() Width()
	var st *StatusTerminal = NewStatusTerminalFd(
		noFieldp,
		STDefaultFd,
		&stderrBuffer,
		&stdoutBuffer,
		NoPrintf, NoPrintf,
	)

	// enable ANSI override to have status during test
	//	- SetTerminal()
	//	- when tests are run, standard error is not a terminal
	isAnsi = st.SetTerminal(ptermx.IsTerminalYes, widthToUse)
	if isAnsi {
		t.Error("isAnsi true")
	}

	// Width()
	width = st.Width()
	if width != widthToUse {
		t.Errorf("bad width %d exp %d", width, widthToUse)
	}

	// NeedsEndStatus()
	needsEndStatus = st.NeedsEndStatus()
	if !needsEndStatus {
		t.Error("needsEndStatus false")
	}

	st.Status(statusValue)
	st.Log(logValue)
	st.LogTimeStamp(logValue)
	st.LogStdout(stdoutValue)
	st.CopyLog(&copyLogBuffer)
	st.Log(copyLogValue)
	st.CopyLog(&copyLogBuffer, ptermx.CopyLogRemove)

	st.EndStatus()

	// “#status \x1b[J”
	//	- “\r\x1b[J#log\n#status \x1b[J”
	//	- “\r\x1b[J250126 12:48:10-08 #log\n#status \x1b[J”
	//	- “\r\x1b[J#status \x1b[J”
	//	- “\r\x1b[J#copyLog\n#status \x1b[J”
	//	- “\n”
	actStatus = stderrBuffer.String()
	statusLines = strings.Split(actStatus, CR)

	// statusLines should be of correct length
	if len(statusLines) != expStatusLines {
		t.Fatalf("bad statusLines length: %d exp %d\n%q",
			len(statusLines), expStatusLines, actStatus,
		)
	}

	// statusLines[0] should have correct prefix
	if !strings.HasPrefix(statusLines[0], statusValue) {
		t.Errorf("statusLines[0] not starting with %q:\n%q",
			statusValue, actStatus,
		)
	}

	// statusLines[1] should contain expected value
	if !strings.Contains(statusLines[1], expStatus1) {
		t.Errorf("statusLines[1] exp %q:\n%q",
			expStatus1, actStatus,
		)
	}

	// statusLines[2] should contain expected value
	if !strings.Contains(statusLines[2], expStatus2) {
		t.Errorf("statusLines[2] exp %q:\n%q",
			expStatus2, actStatus,
		)
	}

	// statusLines[3] should contain expected value
	if !strings.Contains(statusLines[3], expStatus3) {
		t.Errorf("statusLines[3] exp %q:\n%q",
			expStatus3, actStatus,
		)
	}

	// statusLines[4] should contain expected value
	if !strings.Contains(statusLines[4], expStatus4) {
		t.Errorf("statusLines[4] exp %q:\n%q",
			expStatus3, actStatus,
		)
	}

	// statusLines[4] should have newline suffix
	if !strings.HasSuffix(statusLines[4], expStatusSuffix) {
		t.Errorf("statusLines[4] suffix exp %q:\n%q",
			expStatusSuffix, actStatus,
		)
	}

	// stdoutBuffer should contain expected value
	if s := stdoutBuffer.String(); s != expStdout {
		t.Errorf("stdoutBuffer:\n%q exp\n%q", s, expStdout)
	}

	// copyLogBuffer should contain expected value
	if s := copyLogBuffer.String(); s != expCopyLog {
		t.Errorf("copyLogBuffer:\n%q exp\n%q", s, expCopyLog)
	}
}
