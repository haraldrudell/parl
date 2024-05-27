/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pexec provides streaming, context-cancelable system command execution
package pexec

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

// ExecStreamFull executes a system command using the exec.Cmd type and flexible streaming
//   - ExecStreamFull makes streaming [exec.Cmd] easy to use
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
//     However, on process error exit, stderr output is available in the returned error value
//   - — ExecStreamFull does not close stderr but its use has ceased upon return
//   - env: an environment for the process.
//     If env is nil, the new process uses the current process’ environment
//   - ctx: a context that can be used to terminate the process using SIGKILL.
//     ctx nil is panic
//   - startCallback: an optional callback invoked immediately after process start [exec.Cmd.Start]
//   - extraFiles: an optional list of streams for the process’ file descriptors 3…
//   - statusCode: the process’ exit code:
//   - — 0 on successful process termination
//   - — [TerminatedBySignal] -1 on process terminated by signal such as ^C SIGINT or SIGTERM
//   - — otherwise, a command-specific exit value
//   - isCancel: true if context cancel terminated the process with SIGKILL or
//     a stream-copying error occurred
//   - err: any occurring error
//   - upon return from ExecStreamFull there are 5 outcomes:
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
//   - ExecStreamFull blocks during command execution until:
//   - — the started sub-process terminates
//   - — ctx context is canceled causing the process to be terminated by signal [unix.SIGKILL]
//   - — the process’ context is canceled due to an error ocurring in stream-copying threads
//   - ExecStreamFull may fail prior to [exec.Cmd.Wait] command-execution:
//   - — args does not contain a valid command
//   - — an error occured during stream-copying thread creation
//   - — [exec.Cmd.Start] returned an error occuring prior to process start
//   - Up to 3 stream-copying threads are used to copy data when stdin, stdout or stderr are non-nil
//     and not os.Stdin, os.Stdout or os.Stderr respectively
//   - for stdout and stderr, the [pio] package has usable types:
//   - — [pio.NewWriteCloserToString]
//   - — [pio.NewWriteCloserToChan]
//   - — [pio.NewWriteCloserToChanLine]
//   - — [pio.NewReadWriteCloserSlice]
//   - [parl.EchoModerator] can be used with ExecStreamFull:
//   - — if system commands slow down or lock-up, too many (dozens) invoking goroutines
//     may cause increased memory consumption, thrashing or exhaust of file handles, ie.
//     an uncontrollable host state
//   - — EchoModerator notifies of slow or hung commands and limits parallelism
func ExecStreamFull(
	stdin io.Reader, stdout io.Writer, stderr io.Writer,
	env []string, ctx context.Context, startCallback StartCallback, extraFiles []*os.File,
	args ...string,
) (statusCode int, isCancel bool, err error) {

	// prepare the [exec.Cmd] process-start structure in execCmd
	if len(args) == 0 {
		err = perrors.ErrorfPF("%w", ErrArgsListEmpty)
		return
	}
	// execCtx allows for local cancel, ie. failing copyThreads
	var execCtx = parl.NewCancelContext(ctx)
	// get Cmd structure, possibly resolve args[0] using environment PATH
	var execCmd *exec.Cmd = exec.CommandContext(execCtx, args[0], args[1:]...)
	// possibly replace current process's environment os.Environ()
	if env != nil {
		execCmd.Env = env
	}
	// add extra sub-process file descriptors
	if len(extraFiles) > 0 {
		execCmd.ExtraFiles = extraFiles
	}

	var isDebug = parl.IsThisDebug()
	if isDebug {
		// expensive
		defer parl.Debug("waitgroup and closers complete")
	}

	// sub-process stream management

	// waitgroup making copy-threads awaitable
	var wg sync.WaitGroup
	// thread-safe error container for [io.Copy] errors and runtime panics in copy-threads
	var errs parl.ErrSlice
	// closers are streams to close on failure prior to process start
	//	- once [exec.Cmd.Start] succeeds, [exec.Cmd.Wait] will close these streams
	var closers []io.Closer
	// isStart is true if the process was successfully started
	//	- by [exec.Cmd.Start] returning without error
	var isStart = false
	// possibly close closers, await copy-threads and collect their errors
	defer execStreamFullEnd(&isStart, &closers, &wg, &errs, isDebug, &err)

	// pipe stdin to process
	if stdin != nil {
		if stdin == os.Stdin {
			execCmd.Stdin = stdin
		} else {
			// if not stdin, a copy-thread is required
			var ioWriteCloser io.WriteCloser
			if ioWriteCloser, err = execCmd.StdinPipe(); err != nil {
				err = perrors.ErrorfPF("execCmd.StdinPipe %w", err)
				return // pipe error return
			}
			closers = append(closers, ioWriteCloser)
			wg.Add(1)
			go copyThread("stdin", stdin, ioWriteCloser, &errs, execCtx, &wg)
		}
	}

	// pipe stdout to process
	if stdout != nil {
		if stdout == os.Stdout || stdout == os.Stderr {
			execCmd.Stdout = stdout
		} else {
			// if not stdout, a copy-thread is required
			var ioReadCloser io.ReadCloser
			if ioReadCloser, err = execCmd.StdoutPipe(); err != nil {
				err = perrors.ErrorfPF("execCmd.StdoutPipe %w", err)
				return // pipe error return
			}
			closers = append(closers, ioReadCloser)
			wg.Add(1)
			go copyThread("stdout", ioReadCloser, stdout, &errs, execCtx, &wg)
		}
	}

	// pipe stderr to process
	if stderr != nil {
		if stderr == os.Stdout || stderr == os.Stderr {
			execCmd.Stderr = stderr
		} else {
			// if not stderr, a copy-thread is required
			var ioReadCloser io.ReadCloser
			if ioReadCloser, err = execCmd.StderrPipe(); err != nil {
				err = perrors.ErrorfPF("execCmd.StderrPipe %w", err)
				return // pipe error return
			}
			closers = append(closers, ioReadCloser)
			wg.Add(1)
			go copyThread("stderr", ioReadCloser, stderr, &errs, execCtx, &wg)
		}
	}

	parl.Debug("Start")

	// start the process by invoking [exec.Cmd.Start]
	if err = execCmd.Start(); err != nil {
		err = perrors.ErrorfPF("execCmd.Start %w", err)
	} else {
		isStart = true // [exec.Cmd.Start] completed without error
	}
	// invoke startCallback
	if startCallback != nil {
		var e error
		if e = invokeStart(startCallback, execCmd, err); e != nil {
			err = perrors.AppendError(err, perrors.ErrorfPF("startCallback %w", e))
		}
	}
	if err != nil {
		return // command Start error return
	}

	parl.Debug("Wait")

	// because [exeCmd.Wait] immediately closes pipes obtained from:
	//	- [exec.Cmd.StderrPipe] [exec.Cmd.StdoutPipe]
	//	- await those streams being read and written to end
	//	- at the end of this wait, the sub-process is terminated
	wg.Wait()

	// the process has terminated
	//	- Wait will release resources
	err = execCmd.Wait()

	parl.Debug("Wait complete")

	if err == nil {
		return // command completed successfully return
	}
	err = perrors.ErrorfPF("execCmd.Wait %w", err)

	// possible signal terminating the sub-process
	var signal syscall.Signal
	// retrieve statusCode and signal
	_, statusCode, signal, _ = ExitError(err)

	// update isCancel and err for context cancelation
	if execCtx.Err() != nil && // the context was canceled and
		signal == unix.SIGKILL { // the process was terminated by SIGKILL
		// SIGKILL and context cancel: ignore the error, it is caused by context cancelation
		err = nil       // ignore the error caused by context cancelation
		isCancel = true // indicate sub-process terminated by context cancelation
	}

	return // Wait() error return
}

// execStreamFullEnd awaits the end of possible  copy threads
//   - *isStart true: Start was invoked and returned without error.
//     [exec.Cmd.Process] will close the streams in *closers
//   - *isStart false: Start was not invoked or failed.
//     The *closers streams must be closed here
//   - *closers: a list of streams that should be closed
//   - wg: a wait group for possible copying threads
//   - errs: an error container for errors and panics occuring in copy-threads
//   - *isDebug: whether debug printing is active
//   - errp: where errors are aggregated
func execStreamFullEnd(isStart *bool, closers *[]io.Closer, wg parl.Waitable, errs parl.ErrorsSource, isDebug bool, errp *error) {

	// ensure intermediate streams are closed

	// if [exec.Cmd.Start] failed, [exec.Cmd.Process] will not close the streams
	if !*isStart {
		// close streams connecting the sub-process
		for _, c := range *closers {
			if e := c.Close(); e != nil {
				*errp = perrors.AppendError(*errp, perrors.ErrorfPF("stream Close %w", e))
			}
		}
	}

	if isDebug {
		parl.Debug("closers complete")
	}

	// await copy-threads terminations to collect their errors and panics
	wg.Wait()

	// append any copy-thread errors to *errp
	for _, e := range errs.Errors() {
		*errp = perrors.AppendError(*errp, e)
	}

	if isDebug {
		parl.Debug("waitgroup complete")
	}
}

// invokeStart handles any panic occuring in startCallback
func invokeStart(startCallback StartCallback, execCmd *exec.Cmd, e error) (err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	startCallback.StartResult(execCmd, e)

	return
}

// ErrArgsListEmpty is returned when args doe not contain a command
var ErrArgsListEmpty = errors.New("args list empty")
var (
	// [ExecStreamFull] env: use the environment of the parent process
	DefaulEnv []string
	// [ExecStreamFull] startCallback: no startCallback
	NoStartCallback StartCallback
	// [ExecStreamFull] stdin: no stdin
	NoStdin io.Reader
	// [ExecStreamFull] stdout: no stdout
	NoStdout io.Writer
	// [ExecStreamFull] stderr: no stderr
	NoStderr io.Writer
	// [ExecStreamFull] extraFiles: no extraFiles
	NoExtraFiles []*os.File
)

// [ExecStreamFull] startCallback: the signature of startCallback
type StartCallback interface {
	// StartResult is invoked by [ExecStreamFull] unless it fails
	// prior to Start
	//	- StartResult receives the command-description and
	//		process data along with any error occurring during Start
	//	- if err is nil, the command sub-process did start
	//	- StartResult must be thread-safe
	StartResult(execCmd *exec.Cmd, err error)
}
