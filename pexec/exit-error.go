/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"errors"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	TerminatedBySignal = -1
)

var _ = unix.SIGKILL + unix.SIGINT + unix.SIGTERM

// ExitError returns information on why the exec.Cmd.Start process terminated
//   - if hasStatusCode is true, the process terminated with a status code
//   - if hasStatusCode is false, exec.Cmd.Start failed prior to launch
//   - if StatusCode has value -1 or pexec.TerminatedBySignal, the process terminated due to
//     the signal signal. Common signals are:
//   - — unix.SIGINT from ^C
//   - — unix.SIGKILL from context termination
//   - — unix.SIGTERM from operating-system process-termination
func ExitError(err error) (hasStatusCode bool, statusCode int, signal unix.Signal) {
	var exitError *exec.ExitError
	if hasStatusCode = errors.As(err, &exitError); !hasStatusCode {
		return
	}
	if statusCode = exitError.ExitCode(); statusCode != TerminatedBySignal {
		return
	}
	if waitStatus, ok := exitError.ProcessState.Sys().(syscall.WaitStatus); ok {
		signal = waitStatus.Signal()
	}

	return
}
