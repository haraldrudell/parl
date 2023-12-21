/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"errors"
	"io"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	// buffer size if no buffer provided is 1 MiB
	copyContextBufferSize = 1024 * 1024 // 1 MiB
)

// errInvalidWrite means that a write returned an impossible count
//   - cause is buggy [io.Writer] implementation
var ErrInvalidWrite = errors.New("invalid write result")

var ErrCopyShutdown = errors.New("Copy received Shutdown")

var _ = io.Copy

// ContextCopier is an io.Copy cancelable via context
type ContextCopier struct {
	buf []byte

	isShutdown atomic.Bool
	cancelFunc atomic.Pointer[context.CancelFunc]

	// Copy fields

	readCloser  *ContextReader
	writeCloser *ContextWriter
	// g is error channel receiving result from the copying thread
	g parl.GoResult
}

// NewContextCopier copies src to dst aborting if context is canceled
//   - buf is buffer that can be used
//   - if reader implements WriteTo or writer implements ReadFrom,
//     no buffer is required
//   - if a buffer is reqiired ans missing, 1 MiB is allocated
//   - Copy methods does copying
//   - Shutdown method or context cancel aborts Copy in progress
//   - if the runtime type of reader or writer is [io.Closable],
//     a thread is active during copying
func NewContextCopier(buf ...[]byte) (copier *ContextCopier) {
	var c = ContextCopier{}
	if len(buf) > 0 {
		c.buf = buf[0]
	}
	return &c
}

// Copy copies from src to dst until end of data, error, Shutdown or context cancel
//   - Shutdown method or context cancel aborts copy in progress
//   - on context cancel, error returned is [context.Canceled]
//   - on Shutdown, error returned has [ErrCopyShutdown]
//   - if the runtime type of dst or src is [io.Closable],
//     a thread is active during copying
//   - such reader or writer will be closed
func (c *ContextCopier) Copy(
	dst io.Writer,
	src io.Reader,
	ctx context.Context,
) (n int64, err error) {
	if c.readCloser != nil {
		panic(perrors.NewPF("second invocation"))
	} else if dst == nil {
		panic(parl.NilError("dst"))
	} else if src == nil {
		panic(parl.NilError("src"))
	} else if ctx == nil {
		panic(parl.NilError("ctx"))
	}
	// check for shutdown prior to Copy
	if c.isShutdown.Load() {
		err = perrors.ErrorfPF("%w", ErrCopyShutdown)
		return
	}
	defer c.copyEnd(&err) // ensures context to be canceled

	// store context reader writer
	var cancelFunc context.CancelFunc
	ctx, cancelFunc = context.WithCancel(ctx)
	c.cancelFunc.Store(&cancelFunc)
	c.readCloser = NewContextReader(src, ctx)
	c.writeCloser = NewContextWriter(dst, ctx)

	// if either the reader or the writer can be closed,
	// a separate thread is used
	//	- the thread closes in parallel on context cancel forcing an
	//		immediate abort to copying
	if c.readCloser.IsCloseable() || c.writeCloser.IsCloseable() {
		c.g = parl.NewGoResult()
		go c.contextCopierCloserThread(ctx.Done())
	}

	// If the reader has a WriteTo method, use it to do the copy.
	//   - buffer-less one-go copy
	//   - on end of file, err is nil
	//   - err may be read or write errors
	//   - on context cancel, error is [context.Canceled]
	if writerTo, ok := src.(io.WriterTo); ok {
		return writerTo.WriteTo(c.writeCloser) // reader’s WriteTo gets the writer
	}

	// Similarly, if the writer has a ReadFrom method, use it to do the copy
	//   - buffer-less one-go copy
	//   - on end of file, err is nil
	//   - err may be read or write errors
	//   - on context cancel, error is [context.Canceled]
	if readerFrom, ok := dst.(io.ReaderFrom); ok {
		return readerFrom.ReadFrom(c.readCloser) // writer’s ReadFrom gets the reader
	}

	// copy using an intermediate buffer
	//   - on end of file, err is nil
	//   - err may be read or write errors
	//   - on context cancel, error is [context.Canceled]

	// ensure buffer
	var buf = c.buf
	if buf == nil {
		buf = make([]byte, copyContextBufferSize)
	}

	for {

		// read bytes
		var nRead, errReading = c.readCloser.Read(buf)

		// write any read bytes
		if nRead > 0 {
			var nWritten, errWriting = c.writeCloser.Write(buf[:nRead])
			if nWritten < 0 || nRead < nWritten {
				nWritten = 0
				if errWriting == nil {
					errWriting = ErrInvalidWrite
				}
			}

			// handle write outcome
			n += int64(nWritten)
			if errWriting != nil {
				err = errWriting
				return // write error return
			}
			if nRead != nWritten {
				err = io.ErrShortWrite
				return // short write error return
			}
		}

		// handle read outcome
		if errReading == io.EOF {
			return // end of data return
		} else if errReading != nil {
			err = errReading
			return // read error return
		}
	}
}

// Shutdown order the thread to exit and
// wait for its result
//   - every Copy invocation will have a Shutdown
//     either by consumer or the deferred copyEnd method
func (c *ContextCopier) Shutdown() {
	if c.isShutdown.Load() {
		return // already shutdown
	} else if !c.isShutdown.CompareAndSwap(false, true) {
		return // another thread shut down
	}

	// cancel the child context
	//	- any copy in progress is aborted
	//	- if a thread is running, this orders it to exit
	if cfp := c.cancelFunc.Load(); cfp != nil {
		c.cancelFunc.Store(nil)
		(*cfp)() // invoke cancelFunc
	}
}

// ContextCopierCloseThread is used when either the
// reader or the writer is [io.Closable]
//   - on context cancel, the thread closing reader or writer will
//     immediately cancel copying
func (c *ContextCopier) contextCopierCloserThread(done <-chan struct{}) {
	var err error
	defer c.g.SendError(&err)
	defer parl.PanicToErr(&err)

	// wait for thread exit order
	//	- app cancel or ordered to exit
	<-done

	// close reader and writer
	c.close(&err)
}

// copyEnd:
//   - cancels the context,
//   - if thread, awaits the thread to close reader or writer and collects the result
//   - otherwise closes reader and writer
func (c *ContextCopier) copyEnd(errp *error) {

	// ensure Shutdown has been invoked
	if c.isShutdown.Load() {
		*errp = perrors.AppendError(*errp, perrors.ErrorfPF("%w", ErrCopyShutdown))
	} else {
		// cancel the context
		// order any thread to exit
		c.Shutdown()
	}

	// await thread doing close, or do close
	if g := c.g; g.IsValid() {
		// wait for result from thread
		g.ReceiveError(errp)
	} else {
		c.close(errp)
	}
}

// close closes both reader and writer if their runtime type
// implements [io.Closer]
func (c *ContextCopier) close(errp *error) {
	parl.Close(c.readCloser, errp)
	parl.Close(c.writeCloser, errp)
}
