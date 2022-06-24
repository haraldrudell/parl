/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"context"
	"io"
	"os/exec"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// ExecStream executes a system command with flexible streaming
func ExecStream(stdin io.Reader, stdout io.WriteCloser, stderr io.WriteCloser, env []string, ctx context.Context, args ...string) (err error) {
	if len(args) == 0 {
		err = perrors.NewPF("args list empty")
		return
	}

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
			return
		}
		for _, c := range closers {
			if e := c.Close(); e != nil {
				err = perrors.AppendError(err, perrors.Errorf("stream Close %w", e))
			}
		}
	}()

	// get Cmd structure, possibly resolve args[0] using environment PATH
	var execCmd *exec.Cmd
	execCmd = exec.CommandContext(ctx, args[0], args[1:]...)

	// possibly replace current process's environment os.Environ()
	if env != nil {
		execCmd.Env = env
	}

	// pipe stdin to process
	if stdin != nil {
		var ioWriteCloser io.WriteCloser
		if ioWriteCloser, err = execCmd.StdinPipe(); err != nil {
			err = perrors.ErrorfPF("execCmd.StdinPipe %w", err)
			return
		}
		wg.Add(1)
		go copyThread("stdin", stdin, ioWriteCloser, errs.AddErrorProc, &wg)
	}

	// pipe stdout to process
	if stdout != nil {
		var ioReadCloser io.ReadCloser
		if ioReadCloser, err = execCmd.StdoutPipe(); err != nil {
			err = perrors.ErrorfPF("execCmd.StdoutPipe %w", err)
			return
		}
		wg.Add(1)
		go copyThread("stdout", ioReadCloser, stdout, errs.AddErrorProc, &wg)
	}

	// pipe stderr to process
	if stderr != nil {
		var ioReadCloser io.ReadCloser
		if ioReadCloser, err = execCmd.StderrPipe(); err != nil {
			err = perrors.ErrorfPF("execCmd.StderrPipe %w", err)
			return
		}
		wg.Add(1)
		go copyThread("stderr", ioReadCloser, stderr, errs.AddErrorProc, &wg)
	}

	// execute
	err = execCmd.Start()
	isStart = true
	if err != nil {
		err = perrors.Errorf("execCmd.Start %w", err)
		return
	}
	if err = execCmd.Wait(); err != nil {
		err = perrors.Errorf("execCmd.Wait %w", err)
		return
	}
	return
}

func copyThread(label string, reader io.Reader, writer io.Writer, addError func(err error), wg *sync.WaitGroup) {
	var err error
	defer wg.Done()
	parl.Recover("copy command i/o "+label, &err, addError)

	_, err = io.Copy(writer, reader)
}
