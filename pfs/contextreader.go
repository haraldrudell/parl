/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pfs

import (
	"context"
	"io"
)

// ContextReader reader terminated by context
type ContextReader struct {
	ctx context.Context
	io.Reader
}

// NewContextReader instantiates ContextReader
func NewContextReader(ctx context.Context, reader io.Reader) io.Reader {
	return &ContextReader{ctx: ctx, Reader: reader}
}

func (cr *ContextReader) Read(p []byte) (n int, err error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.Reader.Read(p)
}
