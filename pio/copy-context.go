/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"io"
)

// CopyContext is like [io.Copy] but is cancelable via context
//   - dst: the writer data is copied to
//   - src: the reader where data is read
//   - buf: a buffer that can be used
//   - buf [pio.NoBuffer]: no buffer is available
//   - — if reader implements WriteTo or writer implements ReadFrom,
//     no buffer is required
//   - — if a buffer is required and missing, 1 MiB is allocated
//   - ctx: context that when canceled, aborts copying
//   - — context cancel is detected on any [io.File.Read] or [io.File.Write] invocation and
//     carried out by parallel [io.File.Close] if either reader or writer is closable
//   - CopyContext closes both reader and writer if their runtime type is closable
//   - may launch 1 thread while copying
//   - err may be [context.Canceled]
func CopyContext(dst io.Writer, src io.Reader, buf []byte, ctx context.Context) (written int64, err error) {
	return NewContextCopier(buf).Copy(dst, src, ctx)
}

var _ = io.Copy
