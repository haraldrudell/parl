/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"io"
)

type ContextReader struct {
	reader io.Reader
	ContextCloser
	ctx context.Context
}

var _ io.ReadCloser = &ContextReader{}

func NewContextReader(reader io.Reader, ctx context.Context) (contextReader *ContextReader) {
	var closer, _ = reader.(io.Closer)
	return &ContextReader{
		reader:        reader,
		ContextCloser: *NewContextCloser(closer),
		ctx:           ctx,
	}
}

func (c *ContextReader) Read(p []byte) (n int, err error) {
	if err = c.ctx.Err(); err != nil {
		return
	}
	return c.reader.Read(p)
}
