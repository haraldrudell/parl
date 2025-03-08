/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"io"
)

// ContextReader is an [io.ReadCloser] that aborts on context cancel
//   - on context cancel, Read returns error [context.Canceled]
//   - If the runtime type of reader implements [io.Close], it ContextReader can close it
type ContextReader struct {
	reader io.Reader // Read()
	// idempotent pannic-free closer if reader implemented [io.Closer]
	//	- Close() IsClosable()
	ContextCloser
	ctx context.Context
}

var _ io.ReadCloser = &ContextReader{}

// NewContextReader returns an [io.ReadCloser] that aborts on context cancel
//   - on context cancel, Read returns error [context.Canceled]
//   - If the runtime type of reader implements [io.Close], it can be closed
func NewContextReader(reader io.Reader, ctx context.Context) (contextReader *ContextReader) {
	var closer, _ = reader.(io.Closer)
	contextReader = &ContextReader{
		reader: reader,
		ctx:    ctx,
	}
	NewContextCloser(closer, &contextReader.ContextCloser)
	return
}

var _ = context.Canceled

// Read is like [io.Reader.Read] but cancels if the context is canceled
//   - on context cancel, the error returned is [context.Canceled]
func (c *ContextReader) Read(p []byte) (n int, err error) {
	if err = c.ctx.Err(); err != nil {
		return
	}
	return c.reader.Read(p)
}
