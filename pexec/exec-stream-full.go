/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pexec provides streaming, context-cancelable system command execution
package pexec

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

type StartCallback func(execCmd *exec.Cmd, err error)

// ExecStreamFull executes a system command using the exec.Cmd type and flexible streaming.
//   - ExecStreamFull blocks during command execution
//   - ExecStreamFull returns any error occurring during launch or execution including
//     errors in copy threads
//   - successful exit is: statusCode == 0, isCancel == false, err == nil
//   - statusCode may be set by the process but is otherwise:
//   - — 0 successful exit
//   - — -1 process was killed by signal such as ^C or SIGTERM
//   - context cancel exit is: statusCode == -1, isCancel == true, err == nil
//   - failure exit is: statusCode != 0, isCancel == false, err != nil
//   - —
//   - args is the command followed by arguments.
//   - args[0] must specify an executable in the file system.
//     env.PATH is used to resolve the command executable
//   - if stdin stdout or stderr are nil, the are /dev/null
//     Additional threads are used to copy data when stdin stdout or stderr are non-nil
//   - os.Stdin os.Stdout os.Stderr can be provided
//   - for stdout and stderr pio has usable types:
//   - — pio.NewWriteCloserToString
//   - — pio.NewWriteCloserToChan
//   - — pio.NewWriteCloserToChanLine
//   - — pio.NewReadWriteCloserSlice
//   - any stream provided is not closed. However, upon return from ExecStream all i/o operations
//     have completed and streams may be closed as the case may be
//   - ctx is used to kill the process (by calling os.Process.Kill) if the context becomes
//     done before the command completes on its own
//   - startCallback is invoked immediately after cmd.Exec.Start returns with
//     its result. To not use a callback, set startCallback to nil
//   - If env is nil, the new process uses the current process’ environment
//   - use ExecStreamFull with parl.EchoModerator
//   - if system commands slow down or lock-up, too many (dozens) invoking goroutines
//     may cause increased memory consumption, thrashing or exhaust of file handles, ie.
//     an uncontrollable host state
func ExecStreamFull(stdin io.Reader, stdout io.WriteCloser, stderr io.WriteCloser,
	env []string, ctx context.Context, startCallback StartCallback, extraFiles []*os.File,
	args ...string) (statusCode int, isCancel bool, err error) {
	if len(args) == 0 {
		err = perrors.ErrorfPF("%w", ErrArgsListEmpty)
		return
	}

	// execCtx allows for local cancel, ie. failing copyThreads
	var execCtx = parl.NewCancelContext(ctx)

	// thread management: waitgroup and thread-safe error store
	var wg sync.WaitGroup
	var errs parl.ErrSlice
	var isDebug = parl.IsThisDebug()
	if isDebug {
		// expensive
		defer parl.Debug("waitgroup and closers complete")
	}
	// close if we are aborting
	var closers []io.Closer
	var isStart = false
	defer execStreamFullEnd(&isStart, &closers, &wg, &errs, isDebug, &err)
	defer func() {
		if isStart {
			return // do nothing: if exec.Cmd.Start succeeded, exe.Cmd close the streams
		}
		for _, c := range closers {
			if e := c.Close(); e != nil {
				err = perrors.AppendError(err, perrors.ErrorfPF("stream Close %w", e))
			}
		}
	}()

	// get Cmd structure, possibly resolve args[0] using environment PATH
	var execCmd *exec.Cmd = exec.CommandContext(execCtx, args[0], args[1:]...)

	// possibly replace current process's environment os.Environ()
	if env != nil {
		execCmd.Env = env
	}

	// pipe stdin to process
	if stdin != nil {
		if stdin == os.Stdin {
			execCmd.Stdin = stdin
		} else {
			var ioWriteCloser io.WriteCloser
			if ioWriteCloser, err = execCmd.StdinPipe(); err != nil {
				err = perrors.ErrorfPF("execCmd.StdinPipe %w", err)
				return // pipe error return
			}
			wg.Add(1)
			go copyThread("stdin", stdin, ioWriteCloser, &errs, execCtx, &wg)
		}
	}

	// pipe stdout to process
	if stdout != nil {
		if stdout == os.Stdout || stdout == os.Stderr {
			execCmd.Stdout = stdout
		} else {
			var ioReadCloser io.ReadCloser
			if ioReadCloser, err = execCmd.StdoutPipe(); err != nil {
				err = perrors.ErrorfPF("execCmd.StdoutPipe %w", err)
				return // pipe error return
			}
			wg.Add(1)
			go copyThread("stdout", ioReadCloser, stdout, &errs, execCtx, &wg)
		}
	}

	// pipe stderr to process
	if stderr != nil {
		if stderr == os.Stdout || stderr == os.Stderr {
			execCmd.Stderr = stderr
		} else {
			var ioReadCloser io.ReadCloser
			if ioReadCloser, err = execCmd.StderrPipe(); err != nil {
				err = perrors.ErrorfPF("execCmd.StderrPipe %w", err)
				return // pipe error return
			}
			wg.Add(1)
			go copyThread("stderr", ioReadCloser, stderr, &errs, execCtx, &wg)
		}
	}

	if len(extraFiles) > 0 {
		execCmd.ExtraFiles = extraFiles
	}

	// execute
	parl.Debug("Start")
	if err = execCmd.Start(); err != nil {
		err = perrors.ErrorfPF("execCmd.Start %w", err)
	}
	isStart = true
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
	if err = execCmd.Wait(); err != nil {
		err = perrors.ErrorfPF("execCmd.Wait %w", err)
	}
	parl.Debug("Wait complete")
	if err != nil {
		var hasStatusCode bool
		var signal syscall.Signal
		hasStatusCode, statusCode, signal, _ = ExitError(err)

		// was the context canceled?
		if execCtx.Err() != nil &&
			hasStatusCode && // there was an exec.ExitError
			statusCode == TerminatedBySignal && // the process was terminated by a signal
			signal == unix.SIGKILL { // in fact SIGKILL
			// if it was SIGKILL, ignore it: it was caused by context cancelation
			err = nil // ignore the error
			isCancel = true
		}

		return // Wait() error return
	}
	return // command completed successfully return
}

func execStreamFullEnd(isStart *bool, closers *[]io.Closer, wg *sync.WaitGroup, errs parl.ErrorsSource, isDebug bool, errp *error) {
	if *isStart {
		return // do nothing: if exec.Cmd.Start succeeded, exe.Cmd close the streams
	}
	for _, c := range *closers {
		if e := c.Close(); e != nil {
			*errp = perrors.AppendError(*errp, perrors.ErrorfPF("stream Close %w", e))
		}
	}
	if isDebug {
		parl.Debug("closers complete")
	}
	for _, e := range errs.Errors() {
		*errp = perrors.AppendError(*errp, e)
	}
	wg.Wait()
	if isDebug {
		parl.Debug("waitgroup complete")
	}
}

func invokeStart(startCallback StartCallback, execCmd *exec.Cmd, e error) (err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	startCallback(execCmd, e)

	return
}
