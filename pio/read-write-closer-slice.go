/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"bufio"
	"io"
	"io/fs"
	"slices"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

var _ bufio.ReadWriter

// ReadWriteCloserSlice is a read-writer with
// slice as intermediate storage
//   - may copy directly between write and read threads
//   - closable [bufio.ReadWriter]
//   - initialization free, thread-safe
type ReadWriteCloserSlice struct {
	// makes Read critical section
	//	- parallel Read invocations require critical section
	//	- Read blocks on readerWait while holding readLock
	//	- separate because Write Close must be allowed during Read
	readLock parl.Mutex
	// stateLock makes data readerP readerN readerWait thread-safe
	//	- no thread blocks while holding stateLock
	//	- accessed by readEvent Write Buffer
	//	- atomizes isClosed and readerWait in readEvent
	//	- atomizes isClosed and readerWait in Close
	//	- atomized readerWait operations in Write
	stateLock parl.Mutex
	// data contains buffered data written by Write
	// not yet read by Read
	//	- slice-away slice
	//	- behind stateLock
	data, data0 []byte
	// readerP is available buffer from a reader
	//	- behind stateLock
	//	- Write writes to it if non-nil
	readerP []byte
	// readerN is number of bytes written to readerP
	// by Write
	//	- written behind stateLock
	readerN parl.Atomic64[int]
	// readerWait allows a reader-thread to wait
	// for Write or Close when the slice is empty
	//	- opened and closed behind stateLock
	//	- opened by Read behind stateLock
	//	- closed by Close behind stateLock
	//	- closed by Write behind stateLock
	readerWait parl.CyclicAwaitable
	// isClosed is true when readwriter is closed
	//	- read may be in deferred close, ie. still readable
	//	- written behind stateLock
	//	- Close must have lock for atomizing isClosed and readerWait.
	//		Because lock provides thread-wait, a single atomic can replace Awaitable
	isClosed atomic.Bool
}

// [ReadWriteCloserSlice] is [io.ReadWriteCloser]
var _ io.ReadWriteCloser = &ReadWriteCloserSlice{}

// NewReadWriteCloserSlice returns an object that copies from Write to Read
// and is closable
func NewReadWriteCloserSlice(fieldp ...*ReadWriteCloserSlice) (readWriteCloser *ReadWriteCloserSlice) {

	// get readWriteCloser
	if len(fieldp) > 0 {
		readWriteCloser = fieldp[0]
	}
	if readWriteCloser == nil {
		readWriteCloser = &ReadWriteCloserSlice{}
	}

	return
}

// Write copies data directly to Reader buffer or to intermediate data slice
//   - p: buffer to write
//   - n: number of bytes written.
//     If err is nil, n == len(p)
//   - err: if after Close, returns ErrFileAlreadyClosed
//   - — check using: errors.Is(err, fs.ErrClosed)
//   - — no other errors
//   - thread-safe
func (r *ReadWriteCloserSlice) Write(p []byte) (n int, err error) {

	// fast check outside lock
	if r.isClosed.Load() {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
		return // closed error-return
	}
	var remaining = len(p)
	// remaining bytes to write
	if remaining == 0 {
		return
	}
	defer r.stateLock.Lock().Unlock()

	// Write to closed writer is error
	//	- inside lock check
	if r.isClosed.Load() {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
		return // closed error-return
	}

	// hand off:
	//	- by direct copy to reader buffer or
	//	- to r.data

	// copy directly to any reader buffer
	//	- if readerWait open, a Read invocation is waiting
	//	- ecause writeLock is held, Close cannot happen
	if !r.readerWait.IsClosed() && r.readerP != nil {
		var nToReaderP = min(len(p), len(r.readerP))
		// transfer to reader
		copy(r.readerP[:nToReaderP], p)
		r.readerN.Store(nToReaderP)
		r.readerP = nil
		// upate send buffer
		p = p[nToReaderP:]
		n += nToReaderP
		r.readerWait.Close()
		if remaining -= nToReaderP; remaining == 0 {
			return // all written return
		}
	}
	// there is data for buffer

	// append the rest to data
	pslices.SliceAwayAppend(&r.data, &r.data0, p, parl.NoZeroOut)
	n += len(p)

	return // good write return
}

// Read returns at most len(p) bytes read in n and possibly io.EOF
//   - p: buffrer to read to
//   - n: number of bytes read
//   - err: only [io.EOF]
//   - —
//   - if one or more buffered bytes are available, those are returned
//   - if the slice is empty and closed: EOF
//   - Read blocks if no data is available and Close has not occurred
//   - Read blocks until Write of any data or Close
//   - Read returns io.EOF once buffer is empty and Close has occurred
//   - this implementation only returns the error io.EOF
//   - this implementation returns no data, n == 0, with io.EOF
//   - n may be less than len(p)
//   - thread-safe
func (r *ReadWriteCloserSlice) Read(p []byte) (n int, err error) {

	// check for empty buffer
	if len(p) == 0 {
		return // empty read return
	}
	// critical section for Read invocations
	defer r.readLock.Lock().Unlock()

	// read from data or announce reader buffer
	if n, err = r.readerEvent(p); n > 0 || err != nil {
		return // bytes read from r.data or io.EOF return
	}
	// Write can now transfer data to p

	// await close or write
	<-r.readerWait.Ch()
	// Close or non-empty Write happened

	// collect any direct-copy by Write invocation
	if n = r.readerN.Load(); n > 0 {
		return // direct copy return
	}

	// check for EOF
	if r.isClosed.Load() {
		err = io.EOF
		return
	}

	panic(perrors.NewPF("Reader dead end"))
}

// Close prevents further Writes and causes Read to evetually return io.EOF
//   - err: nil on first invocation, then ErrClosed
//   - —
//   - subsequent Close returns ErrFileAlreadyClosed
//   - check using: errors.Is(err, fs.ErrClosed)
//   - thread-safe
func (r *ReadWriteCloserSlice) Close() (err error) {

	// outside lock check
	if r.isClosed.Load() {
		return // already closed error return
	}
	// atomize isClosed.Close and readerWait.Close
	defer r.stateLock.Lock().Unlock()

	// inside lock check
	if r.isClosed.Load() {
		return // not winner thread
	}
	// this thread is winner thread

	// close
	r.isClosed.Store(true)

	// signal close to any waiting reader
	r.readerWait.Close()

	return
}

// Buffer returns a clone of current data
func (r *ReadWriteCloserSlice) Buffer() (buffer []byte) {
	defer r.stateLock.Lock().Unlock()

	buffer = slices.Clone(r.data)
	return
}

// readerEvent reads fom data slice or announce a buffer for Write invocations
//   - p: buffer to read to, cannot be zero-length
//   - n: non-zero if buffer data was not empty and bytes read from it
//   - err: non-zero if no buffered data and Close had been invoked
//   - —
//   - — if any buffered data is available, that is returned
//   - otherwise, readerWait is armed and Write can access p
func (r *ReadWriteCloserSlice) readerEvent(p []byte) (n int, err error) {
	defer r.stateLock.Lock().Unlock()

	// collect any buffered data
	if len(r.data) > 0 {
		n = min(len(r.data), len(p))
		copy(p[:n], r.data)
		r.data = r.data[n:]
		return // buffered data was read return: n > 0
	}

	// check for EOF
	if r.isClosed.Load() {
		err = io.EOF
		return // empty and EOF return: err io.EOF
	}

	// flag buffer available
	r.readerP = p
	r.readerN.Store(0)
	r.readerWait.Open()

	return // empty, not closed slice return: n 0, err nil
}
