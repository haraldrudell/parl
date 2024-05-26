/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"os"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type ThreadSafeRwc struct {
	// slice is a thread-safe buffer
	slice parl.AwaitableSlice[byte]
	// writeLock makes [ThreadSafeRwc.Write] and [ThreadSafeRwc.Close] critical section
	writeLock sync.Mutex
	// readLock makes [ThreadSafeRwc.Read] critical section
	//	- makes readSlice thread-safe
	readLock sync.Mutex
	// temporary buffer for Read
	//	- behind readLock
	readSlice []byte
	// readAlert allows to wake up a waiting Read invocation
	readAlert parl.CyclicAwaitable
	// closeAwait is a thread-safe awaitable semaphore for Close
	closeAwait parl.Awaitable
}

func NewThreadSafeRwc() (readWriteCloser io.ReadWriteCloser) { return &ThreadSafeRwc{} }

// Write saves p to internal buffer
//   - err non-nil: Close was invoked, errorr is os.ErrClosed
//   - — n is 0
//   - err == nil: p was stored in internal buffer, n == len(p)
//   - — no short Write
//   - no error other than os.ErrClosed
//   - thread-safe
func (r *ThreadSafeRwc) Write(p []byte) (n int, err error) {
	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	// error if Close
	if r.closeAwait.IsClosed() {
		err = perrors.ErrorfPF("%w", os.ErrClosed)
		return
	}

	// add to slice
	r.slice.SendClone(p)
	n = len(p)

	// alert reader
	r.readAlert.Close()

	return
}

// Read reads bytes until [io.EOF]
//   - n: number of bytes read. non-zero if not error
//   - err: only error is io.EOF. n is 0
//   - thread-safe
func (r *ThreadSafeRwc) Read(p []byte) (n int, err error) {
	r.readLock.Lock()
	defer r.readLock.Unlock()

	if len(p) == 0 {
		if r.closeAwait.IsClosed() {
			err = io.EOF
		}
		return // Read empty noop return
	}

	// until data or EOF loop
	for {

		// arm recyclable before checking for data
		var _, readAlertCh = r.readAlert.Open()

		// fetch from readSlice
		if len(r.readSlice) > 0 {
			var nx = min(len(p), len(r.readSlice))

			// add to p
			copy(p[:nx], r.readSlice)
			p = p[nx:]
			n += nx

			// remove from readSlice
			if nx == len(r.readSlice) {
				r.readSlice = nil
			} else {
				r.readSlice = r.readSlice[nx:]
			}

			// return on read complete
			if len(p) == 0 {
				return
			}
		}
		// r.readSlice is nil

		// fetch from slices
		//	- may be many slices
		for {

			var slice = r.slice.GetSlice()
			if slice == nil {
				break // r.slice is empty
			}
			var nx = min(len(p), len(slice))

			// add to p
			copy(p[:nx], slice)
			p = p[nx:]
			n += nx

			// save any extra slice bytes
			if nx < len(slice) {
				r.readSlice = slice[nx:]
			}

			// return on read complete
			if len(p) == 0 {
				return
			}
		}
		// r.slice is empty

		// partial read
		//	- if anything was read, return it
		if n > 0 {
			return
		}
		// the reader is empty and no bytes were read

		// check for EOF
		if r.closeAwait.IsClosed() {
			err = io.EOF
			return
		}

		// await Close or Write
		<-readAlertCh
	}
}

// Close: prevents further Write. No errors
//   - idempotent thread-safe
func (r *ThreadSafeRwc) Close() (err error) {
	if r.closeAwait.IsClosed() {
		return
	}
	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	if r.closeAwait.IsClosed() {
		return
	}
	r.closeAwait.Close()
	r.readAlert.Close()

	return
}
