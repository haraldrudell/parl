/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"errors"
	"os/exec"
	"strconv"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

const (
	ExitErrorIncludeStderr = true
	StatusCode1            = 1
)

type ExitErrorData struct {
	Err        error
	ExitErr    *exec.ExitError
	StatusCode int
	Signal     unix.Signal
	Stderr     []byte
}

var _ error = &ExitErrorData{}

// NewExitErrorData returns once-retrieved data on a possible ExitError
//   - if ExitErr field is nil or IsExitError meythod returns false, err does not contain an ExitError
//   - the returned value is an error implementation
func NewExitErrorData(err error, stderr ...[]byte) (exitErrorData *ExitErrorData) {
	var e = ExitErrorData{Err: err}
	if errors.As(err, &e.ExitErr); e.ExitErr != nil {
		_, e.StatusCode, e.Signal, e.Stderr = ExitError(err)
	}
	if len(stderr) > 0 {
		e.Stderr = stderr[0]
	}
	return &e
}

func (e *ExitErrorData) IsExitError() (isExitError bool) {
	return e.ExitErr != nil
}

// IsStatusCode1 returns if the err error chain contains an ExitError
// that indicates status code 1
//   - Status code 1 indicates an unspecified failure of a process
//   - Success has no ExitError and status code 0
//   - Terminated by signal is status code -1
//   - Input syntax error is status code 2
func (e *ExitErrorData) IsStatusCode1() (is1 bool) {
	return e.StatusCode == StatusCode1
}

// IsSignalKill returns true if the err error chain contains an
// ExitError with signal kill
//   - signal kill is the response to a command’s context being
//     canceled
func (e *ExitErrorData) IsSignalKill() (isSignalKill bool) {
	return e.StatusCode == TerminatedBySignal &&
		e.Signal == unix.SIGKILL
}

func (e *ExitErrorData) Error() (exitErrorMessage string) {
	if e.ExitErr != nil {
		exitErrorMessage = e.ExitErr.Error()
	}
	return
}

// ExitErrorString returns a printable string if the err error chain
// contains an ExitError
//   - if no ExitError, the empty string
//   - for non-signal: status code: 1 ‘it fucked up, he got away’
//   - for signal: signal: "abort trap" ‘signal: abort trap’
func (e *ExitErrorData) ExitErrorString(includeStderr ...bool) (errS string) {
	if e.ExitErr != nil {
		if e.StatusCode == TerminatedBySignal {
			errS = "signal: " + strconv.Quote(e.Signal.String())
		} else {
			errS = "status code: " + strconv.Itoa(e.StatusCode)
		}
		errS += "\x20"
	}
	errS += "‘" + perrors.Short(e.Err) + "’"
	if len(includeStderr) > 0 && includeStderr[0] &&
		len(e.Stderr) > 0 {
		errS += " stderr: ‘" + string(e.Stderr) + "’"
	}
	return
}
