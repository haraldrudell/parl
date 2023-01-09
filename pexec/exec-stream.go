/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

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

var ErrArgsListEmpty = errors.New("args list empty")

// ExecStream executes a system command using the exec.Cmd type and flexible streaming.
//   - ExecStream blocks during command execution
//   - ExecStream returns any errors occurring during launch or execution including
//     errors in copy threads
//   - any stream provided is not closed. However, upon return from ExecStream all i/o operations
//     have completed and streams may be closed as the case may be
//   - successful exit is: statusCode == 0, isCancel == false, err == nil
//   - context cancel exit is: statusCode == -1, isCancel == true, err == nil
//   - — statusCode -1 means the process was terminated by signal such as ^C or SIGTERM
//   - failure is: statusCode != 0, isCancel == false, err != nil
//   - —
//   - args is the command followed by arguments.
//   - args[0] must specify an executable in the file system.
//     env.PATH is used to resolve the command executable
//   - if stdin should not be used, pio.EofReader can be used
//   - for stdout and stderr pio has usable types:
//   - — pio.NewWriteCloserToString
//   - — pio.NewWriteCloserToChan
//   - — pio.NewWriteCloserToChanLine
//   - — pio.NewReadWriteCloserSlice
//   - if stdin stdout or stderr are nil, os.Stdin os.Stdout os.Stderr are used.
//     Additional threads are used to copy data when stdin stdout or stderr are non-nil
//   - If env is nil, the new process uses the current process’ environment
//   - ctx is used to kill the process (by calling os.Process.Kill) if the context becomes
//     done before the command completes on its own
//   - startCallback is invoked immediately after cmd.Exec.Start returns with
//     its result. To not use a callback, set startCallback to nil
func ExecStream(stdin io.Reader, stdout io.WriteCloser, stderr io.WriteCloser,
	ctx context.Context, args ...string) (statusCode int, isCancel bool, err error) {
	return ExecStreamFull(stdin, stdout, stderr, nil, ctx, nil, nil, args...)
}

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
	defer wg.Wait()
	var errs perrors.ParlError
	defer func() {
		err = perrors.AppendError(err, errs.GetError())
	}()

	// close if we are aborting
	var closers []io.Closer
	isStart := false
	defer func() {
		if isStart {
			return // do nothing: if exec.Cmd.Start succeeded, exe.Cmd close the streams
		}
		for _, c := range closers {
			if e := c.Close(); e != nil {
				err = perrors.AppendError(err, perrors.Errorf("stream Close %w", e))
			}
		}
	}()

	// get Cmd structure, possibly resolve args[0] using environment PATH
	var execCmd *exec.Cmd
	_ = execCmd
	execCmd = exec.CommandContext(execCtx, args[0], args[1:]...)

	// possibly replace current process's environment os.Environ()
	if env != nil {
		execCmd.Env = env
	}

	// pipe stdin to process
	if stdin != nil {
		var ioWriteCloser io.WriteCloser
		if ioWriteCloser, err = execCmd.StdinPipe(); err != nil {
			err = perrors.ErrorfPF("execCmd.StdinPipe %w", err)
			return // pipe error return
		}
		wg.Add(1)
		go copyThread("stdin", stdin, ioWriteCloser, errs.AddErrorProc, execCtx, &wg)
	}

	// pipe stdout to process
	if stdout != nil {
		var ioReadCloser io.ReadCloser
		if ioReadCloser, err = execCmd.StdoutPipe(); err != nil {
			err = perrors.ErrorfPF("execCmd.StdoutPipe %w", err)
			return // pipe error return
		}
		wg.Add(1)
		go copyThread("stdout", ioReadCloser, stdout, errs.AddErrorProc, execCtx, &wg)
	}

	// pipe stderr to process
	if stderr != nil {
		var ioReadCloser io.ReadCloser
		if ioReadCloser, err = execCmd.StderrPipe(); err != nil {
			err = perrors.ErrorfPF("execCmd.StderrPipe %w", err)
			return // pipe error return
		}
		wg.Add(1)
		go copyThread("stderr", ioReadCloser, stderr, errs.AddErrorProc, execCtx, &wg)
	}

	if len(extraFiles) > 0 {
		execCmd.ExtraFiles = extraFiles
	}

	// execute
	err = execCmd.Start()
	isStart = true
	if startCallback != nil {
		parl.RecoverInvocationPanic(func() {
			startCallback(err)
		}, &err)
	}
	if err != nil {
		err = perrors.Errorf("execCmd.Start %w", err)
		return // command Start error return
	}

	if err = execCmd.Wait(); err != nil {

		// get special exec.ExitError
		var exitError *exec.ExitError
		errors.As(err, &exitError)

		// get status code
		if exitError != nil {
			statusCode = exitError.ExitCode()
		}

		// was the context canceled?
		if execCtx.Err() != nil {
			// if the command did not complete successfully, it’s exec.ExitError
			if exitError != nil {
				// if it was SIGKILL, ignore it: it was cuased by context cancelation
				if waitStatus, ok := exitError.ProcessState.Sys().(syscall.WaitStatus); ok {
					if waitStatus.Signal() == unix.SIGKILL {
						err = nil // ignore the error
						isCancel = true
					}
				}
			}
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

	_, err = io.Copy(writer, reader)
}
