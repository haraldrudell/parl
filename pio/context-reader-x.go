/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"io"
)

// ContextReader reader terminated by context
type ContextReaderX struct {
	ctx context.Context
	io.Reader
}

// NewContextReader instantiates ContextReader
func NewContextReaderX(ctx context.Context, reader io.Reader) io.Reader {
	return &ContextReaderX{ctx: ctx, Reader: reader}
}

func (cr *ContextReaderX) Read(p []byte) (n int, err error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.Reader.Read(p)
}
