/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"errors"
	"io"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
	copyContextBufferSize = 1024 * 1024 // 1 MiB
)

// errInvalidWrite means that a write returned an impossible count.
var ErrInvalidWrite = errors.New("invalid write result")

type ContextCopier struct {
	reader        io.Reader
	readCloser    io.ReadCloser
	writerTo      io.WriterTo
	writer        io.Writer
	writeCloser   io.WriteCloser
	readerFrom    io.ReaderFrom
	hasCloseables bool
	buf           []byte
	errCh         chan error
	endCh         chan struct{}
	ctx           context.Context
}

func NewContextCopier(dst io.Writer, src io.Reader, buf []byte, ctx context.Context) (copier *ContextCopier) {
	if dst == nil {
		panic(perrors.NewPF("dst cannot be nil"))
	}
	if src == nil {
		panic(perrors.NewPF("src cannot be nil"))
	}
	var cReader = NewContextReader(src, ctx)
	var cWriter = NewContextWriter(dst, ctx)

	var c = ContextCopier{
		readCloser:    cReader,
		writeCloser:   cWriter,
		hasCloseables: cReader.IsCloseable() || cWriter.IsCloseable(),
		buf:           buf,
		errCh:         make(chan error, 1),
		endCh:         make(chan struct{}),
		ctx:           ctx,
	}
	c.writerTo, _ = src.(io.WriterTo)
	c.readerFrom, _ = dst.(io.ReaderFrom)
	return &c
}

func (c *ContextCopier) Configuration() (
	hasCloseables,
	hasWriterTo,
	hasReaderFrom bool,
) {
	hasCloseables = c.hasCloseables
	hasWriterTo = c.writerTo != nil
	hasReaderFrom = c.readerFrom != nil
	return
}

func (c *ContextCopier) ContextThread() {
	var err error
	parl.SendErr(c.errCh, &err)
	defer parl.PanicToErr(&err)

	select {
	case <-c.ctx.Done():
	case <-c.endCh:
	}
	if r := c.readCloser; r != nil {
		parl.Close(r, &err)
	}
	if w := c.writeCloser; w != nil {
		parl.Close(w, &err)
	}
}

func (c *ContextCopier) WriteTo() (n int64, err error) {
	return c.writerTo.WriteTo(c.writeCloser)
}

func (c *ContextCopier) ReadFrom() (n int64, err error) {
	return c.readerFrom.ReadFrom(c.readCloser)
}

// copy using buffer
func (c *ContextCopier) BufCopy() (written int64, err error) {

	// ensure buffer
	var buf = c.buf
	if buf == nil {
		buf = make([]byte, copyContextBufferSize)
		c.buf = buf
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
			written += int64(nWritten)
			if errWriting != nil {
				err = errWriting
				return
			}
			if nRead != nWritten {
				err = io.ErrShortWrite
				return
			}
		}

		// handle read outcome
		if errReading == io.EOF {
			return
		} else if errReading != nil {
			err = errReading
			return
		}
	}
}

func (c *ContextCopier) ShutdownThread(errp *error) {
	close(c.endCh)
	parl.CollectError(c.errCh, errp)
}

func (c *ContextCopier) Close(errp *error) {
	parl.Close(c.readCloser, errp)
	parl.Close(c.writeCloser, errp)
}
