/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pexec

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

var ErrArgsListEmpty = errors.New("args list empty")

// ExecStream executes a system command using the exec.Cmd type and flexible streaming.
//   - ExecStream blocks during command execution
//   - ExecStream returns any error occurring during launch or execution including
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
//   - use ExecStream with parl.EchoModerator
//   - if system commands slow down or lock-up, too many (dozens) invoking goroutines
//     may cause increased memory consumption, thrashing or exhaust of file handles, ie.
//     an uncontrollable host state
func ExecStream(stdin io.Reader, stdout io.WriteCloser, stderr io.WriteCloser,
	ctx context.Context, args ...string) (statusCode int, isCancel bool, err error) {
	return ExecStreamFull(stdin, stdout, stderr, nil, ctx, nil, nil, args...)
}

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
	env []string, ctx context.Context, startCallback func(err error), extraFiles []*os.File,
	args ...string) (statusCode int, isCancel bool, err error) {
	if len(args) == 0 {
		err = perrors.ErrorfPF("%w", ErrArgsListEmpty)
		return
	}

	// execCtx allows for local cancel, ie. failing copyThreads
	execCtx := parl.NewCancelContext(ctx)

	// thread management: waitgroup and thread-safe error store
	var wg sync.WaitGroup
	defer parl.Debug("waitgroup complete")
	defer wg.Wait()
	var errs perrors.ParlError
	defer func() {
		err = perrors.AppendError(err, errs.GetError())
	}()

	// close if we are aborting
	var closers []io.Closer
	isStart := false
	defer parl.Debug("closers complete")
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
			go copyThread("stdin", stdin, ioWriteCloser, errs.AddErrorProc, execCtx, &wg)
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
			go copyThread("stdout", ioReadCloser, stdout, errs.AddErrorProc, execCtx, &wg)
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
			go copyThread("stderr", ioReadCloser, stderr, errs.AddErrorProc, execCtx, &wg)
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
		if parl.RecoverInvocationPanic(func() {
			startCallback(err)
		}, &e); e != nil {
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
		hasStatusCode, statusCode, signal = ExitError(err)

		// was the context canceled?
		if execCtx.Err() != nil &&
			hasStatusCode && // there was an exec.ExitError
			statusCode == TerminatedBySignal && // the process was terminated by a signal
			signal == unix.SIGKILL { // in fact SIGKILL
			// if it was SIGKILL, ignore it: it was cuased by context cancelation
			err = nil // ignore the error
			isCancel = true
		}

		return // Wait() error return
	}
	return // command completed successfully return
}

// copyThread copies from a io.Reader to io.Writer.
//   - label is used for thread identification on panics
//   - reader could be the stdin io.Reader being copied to the execCmd.StdinPipe Writer
//   - addError receives panics
//   - on panic exeCtx context is cancelled
//   - the thread itself never fails
func copyThread(label string,
	reader io.Reader, writer io.Writer,
	addError func(err error), execCtx context.Context,
	wg *sync.WaitGroup) {
	defer wg.Done()
	var err error
	defer parl.CancelOnError(&err, execCtx) // cancel the command if copyThread failes
	defer parl.Recover("copy command i/o "+label, &err, addError)

	if _, err = io.Copy(writer, reader); perrors.Is(&err, "%s io.Copy %w", label, err) {

		// if the process terminates quickly, exec.Command might have already closed
		// stdout stderr before the copyThread is scheduled to start
		if errors.Is(err, fs.ErrClosed) {
			err = nil // ignore quickly closed errors
		}

		return
	}
}
