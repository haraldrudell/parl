/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ReadWriteCloserSlice is a read-writer with a slice as intermediate storage. thread-safe.
package pio

import (
	"io"
	"sync"

	"github.com/haraldrudell/parl/perrors"
)

// ReadWriteCloserSlice is a read-writer with a slice as intermediate storage. thread-safe.
//   - Close closes the writer side indicating no further data will be added
//   - Write and Close may return error that can be checked: errors.Is(err, pio.ErrFileAlreadyClosed)
//   - read will eventually return io.EOF after a Close
//   - there are no other errors
type ReadWriteCloserSlice struct {
	dataLock sync.Mutex
	isClosed bool
	data     []byte

	readerCond sync.Cond
}

var _ io.ReadWriteCloser = &ReadWriteCloserSlice{}

func NewReadWriteCloserSlice() (readWriteCloser *ReadWriteCloserSlice) {
	return &ReadWriteCloserSlice{readerCond: *sync.NewCond(&sync.Mutex{})}
}

// Write saves data in slice and returns all bytes written or ErrFileAlreadyClosed
func (wc *ReadWriteCloserSlice) Write(p []byte) (n int, err error) {
	wc.dataLock.Lock()
	defer wc.dataLock.Unlock()

	if wc.isClosed {
		err = perrors.ErrorfPF("%w", ErrFileAlreadyClosed)
		return // closed return
	}

	// consume data
	wc.data = append(wc.data, p...)
	n = len(p)

	return // good write return
}

// Read returns at most len(p) bytes read in n and possibly io.EOF
//   - Read is blocking
//   - n may be less than len(p)
//   - if len(p) > 0, non-error return will have n > 0
func (wc *ReadWriteCloserSlice) Read(p []byte) (n int, err error) {
	wc.readerCond.L.Lock()
	defer wc.readerCond.L.Unlock()

	for {

		var haveData bool
		if haveData, n, err = wc.read(p); haveData || err != nil {
			return // data read or or error return
		}

		// wait for write or close
		wc.readerCond.Wait()
	}
}

func (wc *ReadWriteCloserSlice) read(p []byte) (haveData bool, n int, err error) {
	wc.dataLock.Lock()
	defer wc.dataLock.Unlock()

	// check for EOF or no data
	data := wc.data
	d := len(data)
	if haveData = d > 0; !haveData {
		if wc.isClosed {
			err = io.EOF
			return // eof return: haveData false, err io.EOF
		}
		haveData = len(p) == 0
		return // zero-bytes requested return: haveData true, otherwise haveData false
	}

	// copy one or more bytes
	copy(p, data)

	n = len(p)
	if d <= n {

		// all data consumed
		n = d              // N is bytes read
		wc.data = data[:0] // empty buffer
		return             // all data submitted return
	}

	// only len(p) bytes of data was consumed
	// n already has the shorter len(p) value
	wc.data = data[n:] // remove consumed bytes from data
	return             // p filled return
}

// Close closes thw Write part, may return ErrFileAlreadyClosed
func (wc *ReadWriteCloserSlice) Close() (err error) {
	var doBroadcast bool
	defer func() {
		if doBroadcast {
			wc.readerCond.Broadcast()
		}
	}()

	wc.dataLock.Lock()
	defer wc.dataLock.Unlock()

	if wc.isClosed {
		err = perrors.ErrorfPF("%w", ErrFileAlreadyClosed)
		return // closed return
	}

	wc.isClosed = true
	doBroadcast = true

	return
}
