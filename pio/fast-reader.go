/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

// FastReader is an unbound-buffered reader draining a provided reader quickly
type FastReader struct {
	// reader being drained quickly
	reader io.Reader
	// bufferLock makes bufList thread-safe
	bufferLock parl.Mutex
	// slice-away list of unread buffers
	// - each bufferList element has:
	//	- — len up to 4 KiB
	//	- — cap 4 KiB
	//	- behind bufLock
	bufferList, bufferList0 [][]byte
	// buffer from a reader
	//	- non-nil while direct-copy active
	//	- behind bufferLock
	waitingReaderP []byte
	// bytes written to readerP from a reader
	//	- non-zero if thread wrote to readerP
	//	- written behind bufferLock
	readerN parl.Atomic64[int]
	// readLock makes [FastReader.Read] critical section
	readLock parl.Mutex
	// allows reader threads to wait when no data is available
	readerWait parl.CyclicAwaitable
	// thread-safe error container
	err parl.AtomicError
}

// NewFastReader returns a unbounded buffered reader draining its provided reader quickly
//
// Why:
//   - FastReader is a reader with a thread that immediately reads the entire underlying channel
//     into memory.
//     Basically, [io.ReadAll] in a separate thread
//   - FastReader reads underlying stream to end regardless of Read invocations
//
// Note:
//   - [NewFastReader] new method starts a goroutine fastReaderDrainThread
//   - [NewFastReader.Read] is blocking but io.EOF is not awaitable
//   - [FastReader.Buffer] re-allocates and returns the entire buffer as a single slice
//   - because there is a goroutine, FastReader is heap-allocated
//   - the [FastReader.Read] returns maximum-length byte-slices from their initial allocation
//
// # TODO 260224 deprecate
//
// Design:
//   - [FastReader.Read] is the only operational method
//   - — any available data upon Read is immediately returned
//   - — once the buffer is empty, any error is returned
//   - — errors are: [io.EOF], [os.ErrClosed] or error from the underlying reader
//   - — otherwise, the Read blocks until data or error
//   - the thread exits on error:
//   - — to release resources the underlying reader must be read until error
//   - — Close of the underlying reader is necessary
//   - [FastReader.Buffer] returns a copy of the buffer for testing
//   - [FastReader.Length] returns the number of buffered bytes for testing
//   - storage is slice-of-slices
//   - FastReader features thread-to-thread transfer
//   - uses efficient sliding-window slices
//   - not used 260224
func NewFastReader(reader io.Reader) (fastReader io.Reader) {
	r := FastReader{
		reader:     reader,
		bufferList: make([][]byte, 0, bufListSize),
	}
	go r.fastReaderDrainThread()
	return &r
}

// Read returns bytes from the internal buffer
//   - returns bytes first, errors second
//   - thread-safe
func (r *FastReader) Read(p []byte) (n int, err error) {
	defer r.readLock.Lock().Unlock()

	// read from data or announce reader buffer
	if n, err = r.readerEvent(p); n > 0 || err != nil {
		return // bytes read from r.data or error return
	}
	// thread can now transfer data to p/readerP

	// await thread writing data to p or encountering error
	<-r.readerWait.Ch()
	// non-empty Write or error happened

	// collect any direct-copy bytes by thread
	if n = r.readerN.Load(); n > 0 {
		return // direct copy return
	}

	// it’s error
	if err, _ = r.err.Error(); err == nil {
		err = perrors.NewPF("inconsistency in FastReader")
	}

	return
}

// Length returns the number of bytes curently buffered
//   - thread-safe
func (r *FastReader) Length() (length int) {
	defer r.bufferLock.Lock().Unlock()

	for _, buffer := range r.bufferList {
		length += len(buffer)
	}
	return
}

// Buffer returns a copy of currently buffered data
//   - thread-safe
func (r *FastReader) Buffer() (buffer []byte) {
	defer r.bufferLock.Lock().Unlock()

	var length int
	for _, b := range r.bufferList {
		length += len(b)
	}
	buffer = make([]byte, 0, length)
	for _, b := range r.bufferList {
		buffer = append(buffer, b...)
	}
	return
}

// fastReaderDrainThread incokes reader.Read even if no [FastReader.Read] is present
//   - thread exit on error from reader.Read
func (r *FastReader) fastReaderDrainThread() {
	var err error
	defer parl.DeferredErrorSink(&r.err, &err)
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	var p []byte
	for {
		if p == nil {
			p = make([]byte, fastReaderBufSize)
		}

		// blocks here until data or reader closing
		var n, err = r.reader.Read(p)

		// handle incoming data
		if n > 0 {
			p = r.saveBytes(p[:n])
		}

		if err != nil {
			// store prior to triggering possibly awaiting reader
			r.err.AddError(err)
			r.readerWait.Close()
			// err was already stored
			err = nil
			return
		}
	}
}

// saveBuffer saves the contents of p to the internal buffer
//   - p: bytes to save
//   - pOut p: p was not consumed and can be reused
//   - pOut nil: p was consumed and must be reallocated
//   - store p possibly consuming it by:
//   - — direct copy of bytes in p to readerP
//   - — append bytes in p to the last buffer in bufferList
//   - — append p to bufferList
//   - thread-safe
func (r *FastReader) saveBytes(p []byte) (pOut []byte) {
	var p0 = p
	var pSlicing int
	defer r.bufferLock.Lock().Unlock()

	// try off-loading entire or part of p directly to waiting reader
	if !r.readerWait.IsClosed() && r.waitingReaderP != nil {
		// nr is bytes to copy: greater than zero
		var nr = min(len(p), len(r.waitingReaderP))

		// transfer to reader p
		copy(r.waitingReaderP[:nr], p)
		r.readerN.Store(nr)
		r.waitingReaderP = nil
		r.readerWait.Close()

		// update p
		if nr == len(p) {
			// all data was sent to reader
			//	- p is available for reuse
			pOut = p0
			return
		}
		pSlicing = nr
		p = p[nr:]
	}
	// all bytes of p could not be off-loaded to a thread
	//	- p is the bytes remaining to write
	//	- p0 is original p
	//	- pSlicing is number of bytes already written

	// try appending to active buffer
	if activeBuffer := len(r.bufferList) - 1; activeBuffer >= 0 {
		var ab = r.bufferList[activeBuffer]
		if len(p)+len(ab) <= cap(ab) {
			r.bufferList = append(r.bufferList, p)
			// p was not consumed
			pOut = p0
			return
		}
	}

	// append to buflist
	//	- p is consumed
	if pSlicing > 0 {
		// move bytes to beginning of p
		//	- pSlicing bytes were sliced off at beginning
		//	- p0 length includes p length
		//	- copy is a slice slowness
		copy(p0, p)
		p = p0[:len(p)]
	}
	pslices.SliceAwayAppend1(&r.bufferList, &r.bufferList0, p)

	return
}

// readerEvent reads fom data slice or announce a buffer for Write invocations
//   - n: non-zero if data was not empty
//   - err: non-zero if no data and Close was invoked
//   - otherwise, readerWait is armed and Write can access p
func (r *FastReader) readerEvent(p []byte) (n int, err error) {
	r.bufferLock.Lock()
	defer r.bufferLock.Unlock()

	// collect any buffered data
	var remaining = len(p)
	for len(r.bufferList) > 0 {
		var buffer = r.bufferList[0]

		// copy bytes to p
		var nx = min(len(buffer), len(p))
		copy(p[:nx], buffer)
		p = p[nx:]
		n += nx
		remaining -= nx

		// remove bytes from buffer
		if nx < len(buffer) {
			r.bufferList[0] = buffer[nx:]
		} else {
			r.bufferList = r.bufferList[1:]
		}

		// complete check
		if remaining == 0 {
			return
		}
	}
	// all buffered data was read to p

	// if any data was written to p, return it now
	if n > 0 {
		return
	}

	// return any error once all buffered data has been read
	//	- includes io.EOF
	if err, _ = r.err.Error(); err != nil {
		return
	}

	// publish p so the thread can write directly to it
	r.waitingReaderP = p
	r.readerN.Store(0)
	r.readerWait.Open()

	return
}

const (
	// size of internal read buffer
	fastReaderBufSize = 4096
	// initial capacity of bufList
	bufListSize = 10
)
