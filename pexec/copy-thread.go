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

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// copyThread is a goroutine copying from an [io.Reader] to an [io.Writer]
//   - label: printable string used for thread identification on error and panic
//   - reader: the stream being read from, ie. stdin/stdout/stderr
//   - writer: the stream being written to
//   - errorSink: receives possible stream-copy error or runtime panic
//   - execCtx: a context canceled on error or panic
//   - wg: a wait group making the thread awaitable
//   - reader and writer are streams connecting a child process, eg.
//     the stdin io.Reader being copied to the [exec.Cmd.StdinPipe] Writer
func copyThread(
	label string,
	reader io.Reader, writer io.Writer,
	errorSink parl.ErrorSink, execCtx context.Context, wg parl.DoneLegacy,
) {
	defer wg.Done()
	var err error
	// if a copyThread fails, cancel the command sub-process via context
	defer parl.CancelOnError(&err, execCtx)
	defer parl.RecoverAnnotation("copy command i/o "+label, func() parl.DA { return parl.A() }, &err, errorSink)

	// normal operation is reader ends with [io.EOF] and [io.Copy] returns no error
	if _, err = io.Copy(writer, reader); perrors.Is(&err, "%s io.Copy %w", label, err) {
		// if the process terminates quickly, [exec.Cmd.Command] might have already closed
		// stdout stderr before the copyThread is scheduled to start
		//	- this returns a stream already closed error [os.ErrClosed]
		if errors.Is(err, fs.ErrClosed) {
			err = nil // ignore quickly closed errors
		}

		return
	}
}
