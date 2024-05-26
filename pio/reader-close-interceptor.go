/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
)

type ReaderCloseInterceptor struct {
	// the wrapped Reader: Read()
	io.Reader
	// promoted public methods IsClosed() Ch() Close()
	*CloseWait
}

// NewReaderCloseInterceptor…
//   - reader: [io.Reader] or [io.ReadCloser]: the reader with altered Close behavior
func NewReaderCloseInterceptor(reader io.Reader, label string, closeInterceptor CloseInterceptor) (readCloser io.ReadCloser) {
	return &ReaderCloseInterceptor{
		Reader:    reader,
		CloseWait: NewCloseWait(reader, label, closeInterceptor),
	}
}
