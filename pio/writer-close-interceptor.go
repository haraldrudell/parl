/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
)

type WriterCloseInterceptor struct {
	// the wrapped Writer: Write()
	io.Writer
	// promoted public methods IsClosed() Ch() Close()
	*CloseWait
}

// NewReaderCloseInterceptor…
//   - writer: [io.Writer] or [io.WriteCloser] the writer with altered Close behavior
func NewWriterCloseInterceptor(writer io.Writer, label string, closeInterceptor CloseInterceptor) (readCloser io.WriteCloser) {
	return &WriterCloseInterceptor{
		Writer:    writer,
		CloseWait: NewCloseWait(writer, label, closeInterceptor),
	}
}
