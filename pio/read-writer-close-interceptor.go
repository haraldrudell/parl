/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
)

type ReadWriterCloseInterceptor struct {
	// the wrapped File: many methods
	io.ReadWriter
	// promoted public methods IsClosed() Ch() Close()
	*CloseWait
}

// NewReaderCloseInterceptor…
//   - readWriter: [io.ReadWriter] or [io.ReadWriteCloser]: the reader and writer with altered Close behavior
func NewReadWriterCloseInterceptor(readWriter io.ReadWriter, label string, closeInterceptor CloseInterceptor) (fileCloser io.ReadWriteCloser) {
	return &ReadWriterCloseInterceptor{
		ReadWriter: readWriter,
		CloseWait:  NewCloseWait(readWriter, label, closeInterceptor),
	}
}

func (f *ReadWriterCloseInterceptor) Close() (err error) { return f.CloseWait.Close() }
