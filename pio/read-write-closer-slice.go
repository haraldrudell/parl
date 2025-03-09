/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ReadWriteCloserSlice is a read-writer with a slice as intermediate storage. thread-safe.
package pio

import (
	"io"
	"io/fs"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

// ReadWriteCloserSlice is a read-writer with
// slice as intermediate storage. thread-safe.
type ReadWriteCloserSlice struct {
	// makes Write critical section
	writeLock sync.Mutex
	// true when readwriter is closed
	//	- read may be in deferred close
	//	- written behind dataLock
	isClosed atomic.Bool
	// makes Read critical section
	readLock sync.Mutex
	// allows reader threads to wait when no data is available
	readerWait parl.CyclicAwaitable
	// makes data thread-safe
	dataLock sync.Mutex
	// written data not yet read
	//	- slice-away slice
	//	- behind dataLock
	data, data0 []byte
	// buffer from a reader
	//	- behind dataLock
	readerP []byte
	// bytes in readerP from a reader
	//	- written behind dataLock
	readerN parl.Atomic64[int]
}

// [ReadWriteCloserSlice] is [io.ReadWriteCloser]
var _ io.ReadWriteCloser = &ReadWriteCloserSlice{}

// NewReadWriteCloserSlice returns an object that copies from Write to Read
// and has Close
func NewReadWriteCloserSlice(fieldp ...*ReadWriteCloserSlice) (readWriteCloser *ReadWriteCloserSlice) {

	if len(fieldp) > 0 {
		readWriteCloser = fieldp[0]
	}
	if readWriteCloser == nil {
		readWriteCloser = &ReadWriteCloserSlice{}
	} else {
		*readWriteCloser = ReadWriteCloserSlice{}
	}

	return
}

// Write copies data directly to Reader buffer or to intermediate data slice
//   - if after Close, returns ErrFileAlreadyClosed
//   - check using: errors.Is(err, fs.ErrClosed)
//   - this implementation returns no other errors
//   - thread-safe
func (r *ReadWriteCloserSlice) Write(p []byte) (n int, err error) {
	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	// Write to closed writer is error
	if r.isClosed.Load() {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
		return // closed return
	}

	// remaining bytes to write
	var remaining = len(p)
	if remaining == 0 {
		return // nothing to write return
	}

	// hand off:
	//	- by direct copy to reader buffer or
	//	- to r.data
	r.dataLock.Lock()
	defer r.dataLock.Unlock()

	// copy directly to any reader buffer
	if !r.readerWait.IsClosed() && r.readerP != nil {
		var nr = min(len(p), len(r.readerP))
		// transfer to reader
		copy(r.readerP[:nr], p)
		r.readerN.Store(nr)
		r.readerP = nil
		// upate send buffer
		p = p[nr:]
		n += nr
		r.readerWait.Close()
		if remaining -= nr; remaining == 0 {
			return // all written return
		}
	}

	// append the rest to data
	pslices.SliceAwayAppend(&r.data, &r.data0, p, pslices.NoZeroOut)
	n += len(p)

	return // good write return
}

// Read returns at most len(p) bytes read in n and possibly io.EOF
//   - Read blocks if no data is available and Close has not occurred
//   - Read blocks until Write of any data or Close
//   - Read returns io.EOF once buffer is empty and Close has occurred
//   - this implementation only returns the error io.EOF
//   - this implementation returns no data, n == 0, with io.EOF
//   - n may be less than len(p)
//   - thread-safe
func (r *ReadWriteCloserSlice) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return // empty read
	}
	r.readLock.Lock()
	defer r.readLock.Unlock()

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
//   - subsequent Close returns ErrFileAlreadyClosed
//   - check using: errors.Is(err, fs.ErrClosed)
//   - thread-safe
func (r *ReadWriteCloserSlice) Close() (err error) {
	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	// filter for already closed
	if r.isClosed.Load() || !r.isClosed.CompareAndSwap(false, true) {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
		return // closed return
	}

	// signal close to any waiting reader
	r.readerWait.Close()

	return
}

func (r *ReadWriteCloserSlice) Buffer() (buffer []byte) {
	r.dataLock.Lock()
	defer r.dataLock.Unlock()

	buffer = slices.Clone(r.data)
	return
}

// readerEvent reads fom data slice or announce a buffer for Write invocations
//   - n: non-zero if data was not empty
//   - err: non-zero if no data and Close was invoked
//   - otherwise, readerWait is armed and Write can access p
func (r *ReadWriteCloserSlice) readerEvent(p []byte) (n int, err error) {
	r.dataLock.Lock()
	defer r.dataLock.Unlock()

	// collect any buffered data
	if len(r.data) > 0 {
		n = min(len(r.data), len(p))
		copy(p[:n], r.data)
		r.data = r.data[n:]
		return
	}

	// check for EOF
	if r.isClosed.Load() {
		err = io.EOF
		return
	}

	// flag buffer available
	r.readerP = p
	r.readerN.Store(0)
	r.readerWait.Open()

	return
}
