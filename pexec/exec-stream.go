/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"context"
	"io"
)

// ExecStream executes a system command using the [exec.Cmd] type with flexible streaming
//   - blocking, flexible streaming, ability to cancel
//   - — like [ExecStreamFull] but with fewer arguments
//   - — single-statement version of [exec.Cmd] with flexible streaming
//   - args: the command and its arguments. args[0] empty is error.
//     If args[0] does not contain path, the command is resolved using the parent process’ env.PATH
//   - stdin: an [io.Reader] producing the new process’ standard input.
//   - — stdin be [os.Stdin] if the parent process does not use its standard input
//   - — stdin can be [pio.EofReader] which is a standard input that is open but does not provide any input
//   - — If stdin is nil, stdin is /dev/null meaning the sub-process’ standard input is closed which
//     for some commands cause immediate process termination
//   - —ExecStreamFull does not close a non-nil stdin but its use has ceased upon return
//   - stdout: an [io.Writer] receiving the process’ standard output
//   - — stdout can be [os.Stdout] causing the sub-process’ output to appear on the terminal
//   - — If stdout is nil, the sub-process’ output is discarded
//   - — ExecStreamFull does not close stdout but its use has ceased upon return
//   - stderr: an [io.Writer] receiving the process’ standard error
//   - — stderr can be [os.Stderr] causing the sub-process’ output to appear on the terminal
//   - — If stderr is nil, the new process’ error output is discarded.
//   - ctx: a context that can be used to terminate the process using SIGKILL.
//     ctx nil is panic
//   - statusCode: the process’ exit code:
//   - — 0 on successful process termination
//   - — [TerminatedBySignal] -1 on process terminated by signal such as ^C SIGINT or SIGTERM
//   - — otherwise, a command-specific exit value
//   - isCancel: true if context cancel terminated the process with SIGKILL or
//     a stream-copying error occurred
//   - err: any occurring error
//   - upon return from ExecStream there are 5 outcomes:
//   - — successful exit: statusCode == 0, err == nil
//   - — failure exit: statusCode != 0, isCancel == false.
//     statusCode is the command-specific value provided by sub-process exit.
//     err is the value returned by [exec.Cmd.Wait] upon process exit.
//     If statusCode is pexec.TerminatedBySignal -1, the process was terminated by signal.
//     The signal value can be obtained from err using [pexec.ExitError]
//   - — context cancel: isCancel == true, err == nil.
//     statusCode is pexec.TerminatedBySignal -1
//   - — stream copying error: isCancel == true, err != nil
//     statusCode is pexec.TerminatedBySignal -1
//   - — other error condition: err != nil, statusCode == 0
//   - for stdout and stderr pio has usable types:
//   - — pio.NewWriteCloserToString
//   - — pio.NewWriteCloserToChan
//   - — pio.NewWriteCloserToChanLine
//   - — pio.NewReadWriteCloserSlice
//   - [parl.EchoModerator] can be used with ExecStream:
//   - — if system commands slow down or lock-up, too many (dozens) invoking goroutines
//     may cause increased memory consumption, thrashing or exhaust of file handles, ie.
//     an uncontrollable host state
//   - — EchoModerator notifies of slow or hung commands and limits parallelism
func ExecStream(
	stdin io.Reader, stdout io.WriteCloser, stderr io.WriteCloser,
	ctx context.Context, args ...string,
) (statusCode int, isCancel bool, err error) {
	return ExecStreamFull(stdin, stdout, stderr, nil, ctx, nil, nil, args...)
}
