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

type NetConnTap struct {
	net.Conn
	tap *Tap
}

// NewNetConnTap returns a data tap for a bidirectional data stream
//   - data from readWriter.Read is written to reads.Write if non-nil
//   - data written to readWriter.Write is written to writes.Write if non-nil
//   - a non-nil errs receives all errors from delegated Read Write reads and writes
//   - if errs is nil, an error from the reads and writes taps is panic
//   - ReadWriterTap impements idempotent Close
//   - if any of readWriter, reads or writes implements io.Close, they are closed on socketTap.Close
//   - the consumer may invoke socketTap.Close to ensure reads and writes are closed
//   - errors in reads or writes do not affect the socketTap consumer
func NewNetConnTap(conn net.Conn, readsWriter, writesWriter io.Writer, addError parl.AddError) (socketTap io.ReadWriter) {
	if conn == nil {
		panic(parl.NilError("readWriter"))
	}
	socketTap = &NetConnTap{
		Conn: conn,
		tap:  NewTap(readsWriter, writesWriter, addError),
	}
	return
}

func (t *NetConnTap) Read(p []byte) (n int, err error) { return t.tap.Read(t.Conn, p) }

func (t *NetConnTap) Write(p []byte) (n int, err error) { return t.tap.Write(t.Conn, p) }

func (t *NetConnTap) Close() (err error) { return t.tap.Close(t.Conn) }
