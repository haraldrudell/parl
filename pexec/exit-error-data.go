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
	ExitErrorIncludeStderr = true
	StatusCode1            = 1
)

var eeNewlineBytes = []byte("\n")

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
//   - for non-signal: status code: 1 ‘read error’
//   - for signal: signal: "abort trap" ‘signal: abort trap’
//   - prints stderr if includeStderr is ExitErrorIncludeStderr and stderr non-empty
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
