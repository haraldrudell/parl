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
	// the status code of a process terminated by signal
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
func ExitError(err error) (hasStatusCode bool, statusCode int, signal unix.Signal, stderr []byte) {

	// determine if err contains an ExitError
	//	- ExitError is the error returned when a child process
	//		for executing a command was created and failed
	var exitError *exec.ExitError
	if hasStatusCode = errors.As(err, &exitError); !hasStatusCode {
		return // not an ExitError return
	}

	// obtain possibly cached standard error output
	//	- if the Stderr field was not assigned and the process
	//		echoes to standard error and fails, then that output may have been
	//		cached in the ExitError
	if len(exitError.Stderr) > 0 {
		stderr = exitError.Stderr
	}

	// obtain the status code returned by the command
	if statusCode = exitError.ExitCode(); statusCode != TerminatedBySignal {
		return // is not terminated by signal
	}

	// handle terminated by signal
	if waitStatus, ok := exitError.ProcessState.Sys().(syscall.WaitStatus); ok {
		signal = waitStatus.Signal()
	}

	return
}
