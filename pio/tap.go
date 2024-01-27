/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

// Tap returns a socket tap producing two streams of data
// read from and written to a socket
type Tap struct {
	closeWinner   atomic.Bool
	IsClosed      parl.Awaitable
	reads, writes io.Writer
	errs          func(err error)
}

func NewTap(reads, writes io.Writer, errs func(err error)) (tap *Tap) {
	return &Tap{
		reads:  reads,
		writes: writes,
		errs:   errs,
	}
}

func (t *Tap) Read(reader io.Reader, p []byte) (n int, err error) {

	// do delegated Read
	n, err = reader.Read(p)

	// copy read data to reads
	if n > 0 && t.reads != nil {
		var readsN, readsErr = t.reads.Write(p[:n])
		// error in reads
		if readsErr != nil {
			t.handleError(NewPioError(PeReads, readsErr))
		} else if readsN < n {
			// reads short write
			t.handleError(NewPioError(PeReads, io.ErrShortWrite))
		}
	}

	// propagate reader error
	if err != nil && t.errs != nil {
		t.errs(NewPioError(PeRead, err))
	}

	return
}

func (t *Tap) Write(writer io.Writer, p []byte) (n int, err error) {

	// copy data to writes
	if t.writes != nil {
		var writesN, writesErr = t.writes.Write(p[:n])
		// error in writes
		if writesErr != nil {
			t.handleError(NewPioError(PeWrites, writesErr))
		} else if writesN < n {
			// reads short write
			t.handleError(NewPioError(PeWrites, io.ErrShortWrite))
		}
	}

	// do delegated Write
	n, err = writer.Write(p)

	// propagate reader error
	if err != nil && t.errs != nil {
		t.errs(NewPioError(PeWrite, err))
	}

	return
}

func (t *Tap) Close(closer any) (err error) {

	// pick closing invocation
	if !t.closeWinner.CompareAndSwap(false, true) {
		<-t.IsClosed.Ch()
		return
	}
	defer t.IsClosed.Close()

	// close delegate if it implements io.Close
	if closer, ok := closer.(io.Closer); ok {
		if parl.Close(closer, &err); err != nil && t.errs != nil {
			t.errs(NewPioError(PeClose, err))
		}
	}

	// reads and writes
	var e [2]error
	for i, a := range []any{t.reads, t.writes} {
		var closer, ok = a.(io.Closer)
		if !ok {
			continue
		}
		parl.Close(closer, &e[i])
	}

	// handle errors, may panic
	for i, source := range []PIOErrorSource{PeReads, PeWrites} {
		if e[i] == nil {
			continue
		}
		t.handleError(NewPioError(source, e[i]))
	}

	return
}

func (t *Tap) handleError(err error) {
	if t.errs != nil {
		t.errs(err)
	} else {
		panic(err)
	}
}

// MultiWriter creates a writer that duplicates its writes to all the provided writers
//   - func io.MultiWriter(writers ...io.Writer) io.Writer
var _ = io.MultiWriter

// TeeReader returns a Reader that writes to w what it reads from r
//   - func io.TeeReader(r io.Reader, w io.Writer) io.Reader
var _ = io.TeeReader

// Writer is the interface that wraps the basic Write method
var _ io.Writer

// ReadWriter is the interface that groups the basic Read and Write methods.
var _ io.ReadWriter
