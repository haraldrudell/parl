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
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

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
