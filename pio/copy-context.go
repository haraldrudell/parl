/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pio provides a context-cancelable stream copier, a closable buffer, line-based reader
// and other io functions
package pio

import (
	"context"
	"io"
)

// CopyContext is like [io.Copy] but is cancelable via context
//   - CopyContext closes both reader and writer if their runtime type is closable
//   - context cancel is on any Read or Write invocation and by parallel close
//     if either reader or writer is closable
func CopyContext(dst io.Writer, src io.Reader, buf []byte, ctx context.Context) (written int64, err error) {
	return NewContextCopier(buf).Copy(dst, src, ctx)
}

var _ = io.Copy
