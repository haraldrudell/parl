/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"io"
)

// ContextWriter is an [io.WriteCloser] that aborts on context cancel
//   - on context cancel, Write returns error [context.Canceled]
//   - If the runtime type of writer implements [io.Close], it ContextWriter can close it
type ContextWriter struct {
	writer io.Writer // Write()
	// idempotent pannic-free closer if reader implemented [io.Closer]
	//	- Close() IsClosable()
	ContextCloser
	ctx context.Context
}

var _ io.WriteCloser = &ContextWriter{}

// NewContextWriter returns an [io.WriteCloser] that aborts on context cancel
//   - on context cancel, Write returns error [context.Canceled]
//   - If the runtime type of reader implements [io.Close], it can be closed
func NewContextWriter(writer io.Writer, ctx context.Context) (contextWriter *ContextWriter) {
	var closer, _ = writer.(io.Closer)
	return &ContextWriter{
		writer:        writer,
		ContextCloser: *NewContextCloser(closer),
		ctx:           ctx,
	}
}

// Write is like [io.Writer.Write] but cancels if the context is canceled
//   - on context cancel, the error returned is [context.Canceled]
func (c *ContextWriter) Write(p []byte) (n int, err error) {
	if err = c.ctx.Err(); err != nil {
		return
	}
	return c.writer.Write(p)
}
