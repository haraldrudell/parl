/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"errors"
	"io"
	"os"
	"slices"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

// FastWriter is an unbound-buffered write-closer-reader with fast Write
type FastWriter struct {
	// writer wrapped with fast Write
	writer io.Writer
	// writeLock makes [FastWriter.Wrie] critical section
	writeLock sync.Mutex
	// bufferLock makes bufList thread-safe
	bufferLock sync.Mutex
	// slice-away list of unread buffers
	bufferList, bufferList0 [][]byte
	// threadWait allows the thread to wait for data or error
	threadWait parl.CyclicAwaitable
	// makes thread awaitable
	threadAwait parl.Awaitable
	// closeOnce selects Close winner
	closeOnce parl.OnceCh
	// thread-safe error container
	//	- on Close, set to [os.ErrClosed] if not alreay error
	//	- overwritten by any occuring error from writer.Write by thread
	err parl.AtomicError
}

// NewFastWriter returns a unbound-buffered writer providing fast Write
//   - writer: the writer data is eventually written to
//   - [FastWriter.Write] quickly accepts any provided data
//   - — returns any error from previous Write
//   - — data is buffered for deferred writer.Write
//   - [FastWriter.Close] returns any error from previous Write
//   - — prevents further Write
//   - a thread is used to enable Write to return quickly
//   - thread-safe
func NewFastWriter(writer io.Writer) (fastWriter io.WriteCloser) {
	w := FastWriter{
		writer:     writer,
		bufferList: make([][]byte, 0, writeListSize),
	}
	go w.fastWriterThread()
	return &w
}

// var z parl.AtomicMax[uint64]

// Write quickly stores a clone of p in internal buffer
//   - err == nil: n == len(p), no short writes
//   - err !- nil:
//   - — Write after Close returns [os.ErrClosed]
//   - — any error from previous Write invocation returned by writer.Write
//   - because writes are deferred, FastWriter must be closed to determine success
func (w *FastWriter) Write(p []byte) (n int, err error) {
	// var t = time.Now()
	// var max = len(p)
	// if !z.Value(uint64(max)) {
	// 	max = int(z.Max1())
	// }
	// parl.D("Write %d(%d) %s", len(p), max, t.Format(parl.Rfc3339ns))
	// defer func() {
	// 	var t1 = time.Now()
	// 	parl.D("WroteDone %s", t1.Sub(t))
	// }()

	// fast outside lock check
	if err, _ = w.err.Error(); err != nil {
		return // error prior to Write: [os.ErrClose] or error from writer.Write
	} else if len(p) == 0 {
		return // noop write
	}
	defer w.writeLock.Unlock()
	w.writeLock.Lock()

	// return any previosuly occurring error including [os.ErrClosed] from Close
	if err, _ = w.err.Error(); err != nil {
		return // error prior to write
	}
	defer w.bufferLock.Unlock()
	w.bufferLock.Lock()

	// append all data for the thread
	pslices.SliceAwayAppend1(&w.bufferList, &w.bufferList0, slices.Clone(p))
	n = len(p)

	// notify the thread
	w.threadWait.Close()

	return
}

// Close prevents any further Write
//   - err: any previously occuring writer.Write error
//   - idempotent thread-safe
func (w *FastWriter) Close() (err error) {

	if isWinner, done := w.closeOnce.IsWinner(); !isWinner {
		if errNow, _ := w.err.Error(); !errors.Is(errNow, os.ErrClosed) {
			err = errNow // a previously occuring write.Write error
		}
		return
	} else {
		defer done.Done()
	}

	// write [os.ErrClosed] if no error
	var e = perrors.ErrorfPF("%w", os.ErrClosed)
	if _, otherErr := w.err.AddErrorSwap(nil, e); !errors.Is(otherErr, os.ErrClosed) {
		err = otherErr
	}

	// alert thread
	w.threadWait.Close()

	// await thread exit
	<-w.threadAwait.Ch()

	return
}

// Length returns the number of bytes curently buffered
//   - thread-safe
func (w *FastWriter) Length() (length int) {
	w.bufferLock.Lock()
	defer w.bufferLock.Unlock()

	for _, buffer := range w.bufferList {
		length += len(buffer)
	}
	return
}

// Buffer returns a copy of currently buffered data
//   - thread-safe
func (w *FastWriter) Buffer() (buffer []byte) {
	w.bufferLock.Lock()
	defer w.bufferLock.Unlock()

	var length int
	for _, b := range w.bufferList {
		length += len(b)
	}
	buffer = make([]byte, 0, length)
	for _, b := range w.bufferList {
		buffer = append(buffer, b...)
	}
	return
}

// fastWriterThread writes buffered data to writer.Write
//   - thread exit on write error or Close once buffer empty
func (w *FastWriter) fastWriterThread() {
	defer w.threadAwait.Close()
	var err error
	defer parl.DeferredErrorSink(&w.err, &err)
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	// thread alert loop
	for {

		// arm recyclable prior to checking for data
		var _, threadAlertCh = w.threadWait.Open()

		// write any buffered data
		var bufToWrite []byte
		for {
			if bufToWrite = w.buffer(); len(bufToWrite) == 0 {
				break // buffer now empty
			}
			var n int
			if n, err = w.writer.Write(bufToWrite); perrors.IsPF(&err, "writer.Write %w", err) {
				return // writer.Write error return
			} else if n < len(bufToWrite) {
				err = perrors.ErrorfPF("%w", io.ErrShortWrite)
				return // write.Writer short write return
			}
		}
		// the buffer is empty

		// check for Close
		//	- w.err is only written by Close and thread
		if e, _ := w.err.Error(); errors.Is(e, os.ErrClosed) {
			return // closed and end of data return
		}

		// await data or close
		<-threadAlertCh
	}
}

// buffer gets any available buffer
//   - len(buf) > 0: data from b uffer
//   - buf nil: buffer is empty
func (r *FastWriter) buffer() (buf []byte) {
	r.bufferLock.Lock()
	defer r.bufferLock.Unlock()

	if len(r.bufferList) == 0 {
		return
	}
	buf = r.bufferList[0]
	r.bufferList = r.bufferList[1:]

	return
}

const (
	// initial capacity of bufList
	writeListSize = 10
)
