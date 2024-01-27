/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"net"

	"github.com/haraldrudell/parl"
)

type ReadWriterTap struct {
	io.ReadWriter
	t *Tap
}

// NewReadWriterTap returns a data tap for a bidirectional data stream
//   - data from readWriter.Read is written to reads.Write if non-nil
//   - data written to readWriter.Write is written to writes.Write if non-nil
//   - a non-nil errs receives all errors from delegated Read Write reads and writes
//   - if errs is nil, an error from the reads and writes taps is panic
//   - ReadWriterTap impements idempotent Close
//   - if any of readWriter, reads or writes implements io.Close, they are closed on socketTap.Close
//   - the consumer may invoke socketTap.Close to ensure reads and writes are closed
//   - errors in reads or writes do not affect the socketTap consumer
func NewReadWriterTap(readWriter io.ReadWriter, reads, writes io.Writer, errs func(err error)) (socketTap io.ReadWriter) {
	if readWriter == nil {
		panic(parl.NilError("readWriter"))
	}
	socketTap = &ReadWriterTap{
		ReadWriter: readWriter,
		t:          NewTap(reads, writes, errs),
	}
	return
}

func (t *ReadWriterTap) Read(p []byte) (n int, err error) { return t.t.Read(t.ReadWriter, p) }

func (t *ReadWriterTap) Write(p []byte) (n int, err error) { return t.t.Write(t.ReadWriter, p) }

func (t *ReadWriterTap) Close() (err error) { return t.t.Close(t.ReadWriter) }

var _ io.Reader
var _ io.Writer

// Pipe creates a synchronous, in-memory, full duplex network connection
//   - writes are behind lock to unbuffered channel
//   - both returned connections are identical
//   - no threads
//   - func net.Pipe() (net.Conn, net.Conn)
var _ = net.Pipe

// Pipe creates a synchronous in-memory pipe
//   - Write is behind lock and waits for sufficient reads
//   - writer is io.WriteCloser
//   - no threads
//   - func io.Pipe() (*io.PipeReader, *io.PipeWriter)
var _ = io.Pipe
