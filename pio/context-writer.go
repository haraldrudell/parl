/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"io"
)

type ContextWriter struct {
	writer io.Writer
	ContextCloser
	ctx context.Context
}

var _ io.WriteCloser = &ContextWriter{}

func NewContextWriter(writer io.Writer, ctx context.Context) (contextWriter *ContextWriter) {
	var closer, _ = writer.(io.Closer)
	return &ContextWriter{
		writer:        writer,
		ContextCloser: *NewContextCloser(closer),
		ctx:           ctx,
	}
}

func (c *ContextWriter) Write(p []byte) (n int, err error) {
	if err = c.ctx.Err(); err != nil {
		return
	}
	return c.writer.Write(p)
}
