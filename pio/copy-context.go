/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"context"
	"io"
)

// CopyContext is like io.Copy but is cancelable via context
//   - CopyContext closes both reader and writer if they are closable
//   - to cancel a read or write in progress, the reader must be closable
func CopyContext(dst io.Writer, src io.Reader, buf []byte, ctx context.Context) (written int64, err error) {
	var c = NewContextCopier(dst, src, buf, ctx)
	var hasCloseables, hasWriterTo, hasReaderFrom = c.Configuration()

	// ensure either a thread or deferred function close any closable src or dst
	if hasCloseables {
		// a separate thread is used
		//	- the reader or writer or both can be closed
		//	- the thread closes them if the context closes
		go c.ContextThread()
		defer c.ShutdownThread(&err) // oder and wait for thread exit
	} else {
		defer c.Close(&err) // ensure src and dst are both closed if closeable
	}

	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if hasWriterTo {
		return c.WriteTo()
	}

	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if hasReaderFrom {
		return c.ReadFrom()
	}

	return c.BufCopy()
}
