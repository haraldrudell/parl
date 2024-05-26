/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"fmt"
	"io"
	"io/fs"
	"slices"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// TeeWriter is a writer that copies its writes to one or more other writers.
type TeeWriter struct {
	// closeInterceptor replaces regular close and is invoked exactly once on [TeeWriter.Close]
	closeInterceptor TeeWriterCloseInterceptor
	// writers is the list of writers Write is being copied to
	writers []io.Writer
	// isClosed indicates [TeeWriter.Close] was invoked
	isClosed atomic.Bool
}

// TeeWriter returns an [io.WriteCloser] writer that duplicates its Write invocations to
// one or more writers. Close behavior is configurable using [TeeWriterCloseInterceptor]
//   - closeInterceptor: if present determines custom [TeeWriter.Close] behavior
//   - closeInterceptor TeeWriterNoInterceptor: default behavior:
//   - — any writer implementing [io.Closer] is closed
//   - writers: non-empty list of [io.Writer] that the Write method duplicates writes to. Cannot be empty
//   - [TeeWriter.Write] duplicates Write to all writers.
//     Returns [fs.ErrClosed] if after Close, and transparently the first error from any writer
//   - [TeeWriter.Close] closes all writers while honoring closeInterceptor behavior
//     Returns [fs.ErrClosed] if more than one Close invocation
//   - TeeWriter is an [io.MultiWriter] with an optionally custom Close implementation
func NewTeeWriter(closeInterceptor TeeWriterCloseInterceptor, writers ...io.Writer) (teeWriter io.WriteCloser) {
	if len(writers) == 0 {
		panic(perrors.NewPF("Must have one or more writers, writers is empty"))
	}
	for i, w := range writers {
		if w == nil {
			panic(parl.NilError(fmt.Sprintf("writers#%d", i)))
		}
	}
	return &TeeWriter{
		closeInterceptor: closeInterceptor,
		writers:          slices.Clone(writers),
	}
}

// Write duplicates Write to all [io.Write] in writers
func (tw *TeeWriter) Write(p []byte) (n int, err error) {
	// thread-safe check for Close invocation
	if tw.isClosed.Load() {
		// “file already closed” with stack
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
		return // Teewriter is closed return
	}
	for i, writer := range tw.writers {

		// thread-safe check for Close invocation
		if tw.isClosed.Load() {
			// “file already closed” with stack
			err = perrors.ErrorfPF("%w", fs.ErrClosed)
			return // Teewriter is closed return
		}

		// invoke Write for a writer
		n, err = writer.Write(p)
		if err != nil {
			return // Write error return
		} else if n != len(p) {
			// “short write” with stack
			err = perrors.ErrorfPF("writer#%d %w", i, io.ErrShortWrite)
			return // short write return
		}
	}

	return // good write return
}

// Close executes any interceptor and closes all writers
func (w *TeeWriter) Close() (err error) {

	// prevent multiple Close invocations
	if w.isClosed.Load() || !w.isClosed.CompareAndSwap(false, true) {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
		return
	}

	// invoke interceptor if there is one
	var doRegularClose bool
	if doRegularClose, err = w.invokeIntercept(); !doRegularClose {
		return
	}

	// regular Close
	for i, w := range w.writers {
		var closer, ok = w.(io.Closer)
		if !ok {
			continue
		}
		var e error
		parl.Close(closer, &e)
		if e == nil {
			continue
		}
		err = perrors.AppendError(err, perrors.ErrorfPF("writer#%d %w", i, e))
	}

	return
}

// invokeIntercept invokes closeInterceptor if present
func (w *TeeWriter) invokeIntercept() (doRegularClose bool, err error) {
	if doRegularClose = w.closeInterceptor == nil; doRegularClose {
		return
	}
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	return w.closeInterceptor.Closer(w, w.writers)
}

// TeeWriterCloseInterceptor implements a custom Close function for TeeWriter
type TeeWriterCloseInterceptor interface {
	Closer(teeWriter *TeeWriter, writers []io.Writer) (doRegularClose bool, err error)
}

// TeeWriter is similar to [io.MultiWriter]
var _ = io.MultiWriter

// [NewTeeWriter] closer: there is no closer
var TeeWriterNoInterceptor TeeWriterCloseInterceptor

// [TeeWriterCloseInterceptor.Closer] do regular close on return
var TeeWriterDoClose = true

// [TeeWriterCloseInterceptor.Closer] no regular close on return
var TeeWriterNoClose = false
