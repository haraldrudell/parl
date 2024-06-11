/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"os"
)

type FileCloseInterceptor struct {
	// the wrapped File: many methods
	*os.File
	// promoted public methods IsClosed() Ch() Close()
	*CloseWait
}

// NewReaderCloseInterceptor…
//   - readWriter: [io.ReadWriter] or [io.ReadWriteCloser]: the reader and writer with altered Close behavior
func NewFileCloseInterceptor(osFile *os.File, label string, closeInterceptor CloseInterceptor) (fileCloser io.ReadWriteCloser) {
	return &FileCloseInterceptor{
		File:      osFile,
		CloseWait: NewCloseWait(osFile, label, closeInterceptor),
	}
}

func (f *FileCloseInterceptor) Close() (err error) { return f.CloseWait.Close() }
