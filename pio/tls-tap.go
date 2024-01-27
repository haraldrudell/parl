/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"crypto/tls"
	"io"

	"github.com/haraldrudell/parl"
)

type TLSTap struct {
	*tls.Conn
	t *Tap
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
func NewTLSTap(conn *tls.Conn, reads, writes io.Writer, errs func(err error)) (socketTap io.ReadWriter) {
	if conn == nil {
		panic(parl.NilError("conn"))
	}
	socketTap = &TLSTap{
		Conn: conn,
		t:    NewTap(reads, writes, errs),
	}
	return
}

func (t *TLSTap) Read(p []byte) (n int, err error) { return t.t.Read(t.Conn, p) }

func (t *TLSTap) Write(p []byte) (n int, err error) { return t.t.Write(t.Conn, p) }

func (t *TLSTap) Close() (err error) { return t.t.Close(t.Conn) }
