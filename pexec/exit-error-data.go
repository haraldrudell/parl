/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/haraldrudell/parl/pbytes"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

const (
	// [ExitErrorData.ExitErrorString] should include standard error output
	ExitErrorIncludeStderr = true
	// status code 1, which in POSIX means a general error
	StatusCode1 = 1
	// counts [AddStderr] stack frame
	addStderrStack = 1
)

// ExitErrorData provides additional detail on pexec.ExitError values
type ExitErrorData struct {
	// the original error
	Err error
	// Err interpreted as ExitError, possibly nil
	ExitErr *exec.ExitError
	// status code in ExitError, possibly 0
	StatusCode int
	// signal in ExitError, possibly 0
	Signal unix.Signal
	// stderr in ExitError or from argument, possibly nil
	Stderr []byte
}

// ExitErrorData implements error
var _ error = &ExitErrorData{}

// NewExitErrorData returns parse-once data on a possible ExitError
//   - if ExitErr field is nil or IsExitError method returns false, err does not contain an ExitError
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

// IsExitError returns true if an pexec.ExitError is present
//   - false if Err was nil or some other type of error
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
//     canceled. This should be checked together with [context.Context.Err]
//   - SIGKILL can also be sent to the process by the operating system
//     trying to reclaim memory or by other processes
func (e *ExitErrorData) IsSignalKill() (isSignalKill bool) {
	return e.StatusCode == TerminatedBySignal &&
		e.Signal == unix.SIGKILL
}

// the Error method returns the message from any ExitError,
// otherwise empty string
//   - Error also makes ExitErrorData implementing the error
//     interface
func (e *ExitErrorData) Error() (exitErrorMessage string) {
	if e.ExitErr != nil {
		exitErrorMessage = e.ExitErr.Error()
	}
	return
}

// AddStderr adds standard error output at the end of the error message
// for err. Also ensures stack trace.
//   - ExitError has standard error if the Output method was used
//   - NewExitErrorData can also have been provided stderr
func (e *ExitErrorData) AddStderr(err error) (err2 error) {
	if stderr := e.Stderr; len(stderr) > 0 {
		if serr := pbytes.TrimNewline(stderr); len(serr) > 0 {
			err2 = perrors.Errorf("%w stderr: ‘%s’", err, string(serr))
			return // standard error appended to message
		}
	}
	if perrors.HasStack(err) {
		err2 = err
		return // error already has stack trace: no change return
	}
	err2 = perrors.Stackn(err, addStderrStack)
	return // stackk added Stderr return
}

// ExitErrorString returns the ExitError error message and data from
// Err and stderr, not an error value
//   - for non-signal: “status code: 1 ‘read error’”
//   - for signal: “signal: "abort trap" ‘signal: abort trap’”
//   - the error message for err: “message: ‘failure’”
//   - stderr if non-empty from ExitErr or stderr argument and
//     includeStderr is ExitErrorIncludeStderr:
//   - “stderr: ‘I/O error’”
//   - returned value is never empty
func (e *ExitErrorData) ExitErrorString(includeStderr ...bool) (errS string) {
	var s []string
	var stderr []byte
	if e.ExitErr != nil {
		if stderr = e.ExitErr.Stderr; len(stderr) == 0 {
			stderr = e.Stderr
		}

		// it’s either status code or signal
		if e.StatusCode == TerminatedBySignal {
			s = append(s, fmt.Sprintf("signal: %q", e.Signal.String()))
		} else {
			s = append(s, fmt.Sprintf("status code: %d", e.StatusCode))
		}
	}

	// original error message
	s = append(s, fmt.Sprintf("message: ‘%s’", perrors.Short(e.Err)))

	// stderr
	if len(includeStderr) > 0 && includeStderr[0] &&
		len(stderr) > 0 {
		if serr := pbytes.TrimNewline(stderr); len(serr) > 0 {
			s = append(s, fmt.Sprintf("stderr: ‘%s’", string(serr)))
		}
	}

	errS = strings.Join(s, "\x20")
	return
}
