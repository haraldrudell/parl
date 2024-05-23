/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"errors"
	"io"
	"io/fs"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

var cx pruntime.CachedLocation

// CopyThread copies from an io.Reader to an io.Writer.
//   - label is used for thread identification on panics
//   - errCh receives result and makes thread awaitable
//   - if ctx, a CancelContext, is non-nil and error occurs, ctx is cancelled
//   - CopyThread itself never fails
func CopyThread(
	label string, reader io.Reader, writer io.Writer,
	errCh chan<- error, ctx context.Context,
) {
	var err error
	defer func() { errCh <- err }()
	if ctx != nil {
		defer parl.CancelOnError(&err, ctx) // cancel the command if copyThread fails
	}
	defer parl.RecoverAnnotation("copy command i/o "+label, func() parl.DA { return parl.A() }, &err)

	if _, err = io.Copy(writer, reader); perrors.Is(&err, "%s %s %w", label, cx.PackFunc(), err) {

		// if the process terminates quickly, exec.Command might have already closed
		// stdout stderr before the copyThread is scheduled to start
		if errors.Is(err, fs.ErrClosed) {
			err = nil // ignore quickly closed errors
		}

		return
	}
}
