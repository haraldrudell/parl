/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package malib

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pslices"
)

// StdinReader is a reader wrapping the unclosable [os.Stdin.Read]
type StdinReader struct {
	// errorSink is error submitting function
	errorSink parl.ErrorSink1
	// whether error or close has occured in [StdinReader.Read]
	isClosed parl.Awaitable
	// isEOF is true when StdinReader has been closed and read to end
	isEOF atomic.Bool
	// isError is set to true on thread error or panic
	//	- may be nil
	isError *atomic.Bool
	// isActive indicates to thread stdinReader is still operating
	isActive atomic.Bool
	// errCh is error channel from thread
	//	- sends error and panic, max 2
	//	- closes on thread exit
	errCh atomic.Pointer[chan error]
	// bufferLock makes stdinBuffer thread-safe
	bufferLock sync.Mutex
	// stdinBuffer receives byte slices from os.Stdin as they are read
	//	- list of non-empty slices
	//	- behind bufferLock
	stdinBuffer [][]byte
	// dataCh becomes non-emprty when data is added to stdinBuffer
	//	- enables awaiting data from thread
	dataCh chan struct{}
	// readLock makes Read Close critical section
	//	- make sliceList thread-safe
	readLock parl.Mutex
	// sliceList is slice away of local data read from stdinBuffer
	//	- behind readLock
	sliceList, sliceList0 [][]byte
	// isCreatedThread is true when the thread has been created
	//	- written behind readLock
	isCreatedThread atomic.Bool
}

// StdinReader is [io.ReadCloser]
var _ io.ReadCloser = &StdinReader{}

// NewStdinReader returns a error-free reader of standard input that closes on error
//   - errorSink: receives any errors returned by [os.Stdin.Read] or
//     runtime panic in this method.
//   - isError: optional atomic set to true on first error or panic
//     -
//   - [StdinReader.Read] returns bytes read from [os.Stdin] standard input until close or error
//   - [os.Stdin.Read] is blocking and os.Stdin cannot be closed
//   - Therefore, the thread invoking Read may remain until the process exits
//   - StdinReader removes all inband errors and panics and only propagates the fact that
//     no more bytes are available via EOF error
func NewStdinReader(errorSink parl.ErrorSink1, isError ...*atomic.Bool) (stdinReader *StdinReader) {
	stdinReader = &StdinReader{
		errorSink: errorSink,
		dataCh:    make(chan struct{}, 1),
	}
	stdinReader.isActive.Store(true)
	var errCh = make(chan error, errChSize)
	stdinReader.errCh.Store(&errCh)
	if len(isError) > 0 {
		stdinReader.isError = isError[0]
	}
	return
}

// Read reads from standard input
//   - p buffer, max length to read
//   - n: the number of bytes read
//   - err: Read never returns any other error than [io.EOF] on error, panic or close
//   - — subsequent invocations after first EOF receives EOF
//     -
//   - errors and runtime panics are sent to the errorSink or printed to stderr
//   - [os.Stdin] cannot be closed so a blocking [StdinReader.Read] cannot be canceled
//   - the thread invoking Read may remain blocked until process exit
//   - if the stdin pipe is closed by another process,
//     Read keeps blocking but returns on the next keypress.
//     Then, an error os.ErrClosed is sent to errorSink and io.EOF is returned
//   - on process exit, Read is unblocked as stdin is closed
//   - thread-safe
func (r *StdinReader) Read(p []byte) (n int, err error) {
	defer r.readLock.Lock().Unlock()

	// already eof case
	if r.isEOF.Load() {
		err = io.EOF
		return
	} else if !r.isCreatedThread.Load() {
		r.createThread()
	}

	// read thread’s errror channel
	//	- StdinReader does not error
	//	- errors are submitted to errorsSink
	//	- errors and panics will eventually close StdinReader
	r.threadState()

	for {

		// read data from thread and write it to p
		r.readStdinBuffer()
		n = r.copyToP(p)

		// if something was read to p, return it
		if n > 0 {
			return // p partially or completely filled return
		}
		// StdinReader was empty

		// check for EOF
		var errCh = *r.errCh.Load()
		if r.isClosed.IsClosed() || errCh == nil {
			if !r.isEOF.Load() {
				r.isEOF.Store(true)
			}
			err = io.EOF
			return
		}

		// await data Close thread-error
		select {
		case <-r.dataCh:
		case <-r.isClosed.Ch():
			return
		case e, isOpen := <-errCh:
			if !isOpen {
				var errCh chan error
				r.errCh.Store(&errCh)
				continue
			}
			// thread had event
			r.submitThreadError(e)
		}
		// data available or thread event
	}

	// in Go, there is not input/output error,
	//	- it is regular EOF 250404
	//	- eg. sleep 3 | myGoCode
	//
	// if another process closes stdin:
	// os.StdinRead error:
	// “read /dev/stdin: input/output error [*fs.PathError]
	// input/output error [syscall.Errno]”
	// isPanic: false
}

// Close closes StdinReader
//   - always reads thread errors
//   - if stdinBuffer not empty, it is not EOF yet
func (r *StdinReader) Close() (err error) {
	defer r.readLock.Lock().Unlock()

	// mark StdinReader as closed
	r.isClosed.Close()

	// read errors from thread
	if !r.threadState() {
		// thread is still running
		//	- order thread to exit
		if r.isActive.Load() {
			r.isActive.Store(false)
		}
		return // thread is running return
	}
	// thread is exited

	// read any data into local buffer
	r.readStdinBuffer()

	// update EOF
	if len(r.sliceList) == 0 && !r.isEOF.Load() {
		r.isEOF.Store(true)
	}

	return
}

// createThread creates the thread reading from os.Stdin
//   - must be in critical section
func (r *StdinReader) createThread() {

	if r.isCreatedThread.Load() {
		return
	}
	var errCh = *r.errCh.Load()
	go KeystrokesThread(&r.isActive, errCh, &r.stdinBuffer, &r.bufferLock, r.dataCh)
	r.isCreatedThread.Store(true)
}

// readStdinBuffer moves any slices from stdinBuffer to sliceList
//   - sliceCount number of moved slices
//   - —
//   - sliceList is slice-away slice
//   - by emptying stdinBuffer every time,
//     lock contention is reduced and
//     thread does not have to implement slice-away logic.
//     ie. pslices.SliceAwayAppend
func (r *StdinReader) readStdinBuffer() (sliceCount int) {
	r.bufferLock.Lock()
	defer r.bufferLock.Unlock()

	sliceCount = len(r.stdinBuffer)
	if sliceCount == 0 {
		return
	}
	pslices.SliceAwayAppend(&r.sliceList, &r.sliceList0, r.stdinBuffer, parl.DoZeroOut)
	clear(r.stdinBuffer)
	r.stdinBuffer = r.stdinBuffer[:0]

	return
}

// copyToP moves bytes from sliceList to p
//   - p: buffer to fill
//   - n: number of bytes written to p
func (r *StdinReader) copyToP(p []byte) (n int) {

	// while p has room and sliceList has data
	for len(p) > 0 && len(r.sliceList) > 0 {

		// copy from next slice
		var slicep = &r.sliceList[0]
		var n0 = copy(p, *slicep)

		// check for end of data
		if n0 == 0 {
			return
		}

		// update sliceList
		if n0 < len(*slicep) {
			*slicep = (*slicep)[n0:]
		} else {
			r.sliceList = r.sliceList[1:]
		}

		// update n and p
		n += n0
		if n0 == len(p) {
			return
		}
		p = p[n0:]
	}
	return
}

// threadState reads and closes thread error channel
//   - isExit true: thread has exited and its errors have been consumed
//
// there is no thread that is continuously monitoring
// the reader thread
//   - if there was, that would be memory references
//     by the thread
//   - on Read Close invocation,
//     queued-up events can be checked
//
// this thread can set isActive to false
//   - this will cause reader thread to exit on
//     next return press
//
// events from thread to check:
//   - EOF from os.Stdin
//   - error other than EOF from os.Stdin
//   - thread panic
//   - thread exit
func (r *StdinReader) threadState() (isExit bool) {

	// check if thread already exited
	var errCh = *r.errCh.Load()
	if errCh == nil {
		isExit = true
		return
	}

	// non-blocking read from error channel
	//	- on thread exit, errCh will always produce values
	var err error
	var isOpen bool
	for {
		select {
		case err, isOpen = <-errCh:
			if !isOpen {
				var errCh chan error
				r.errCh.Store(&errCh)
				// if thread closed the channel, it is thread-exit
				isExit = true
				return // thread Exit return
			}
		default:
			// there are no queued up errors
			//	- thread is up but not sending error or
			//	- thread is exit and errCh is nil
			return // no queued-up errors return
		}
		// err is StdinErr error from thread

		r.submitThreadError(err)
	}
}

// submitThreadError sends thread error to errorSink
func (r *StdinReader) submitThreadError(err error) {

	// all errors from thread are StdinErr
	var e = err.(*StdinErr)
	if e.RecoverAny != nil {

		// thread had panic
		//	- make recover valu an error with stack trace
		err = parl.EnsureError(e.RecoverAny)
	} else if errors.Is(e.Err, io.EOF) {

		// os.Stdin.Read returned EOF
		//	- close will become EOF once reader is empty
		r.isClosed.Close()
		return // do not submit EOF return
	} else {

		// os.Stdin.Read returned error other than EOF
		err = perrors.ErrorfPF("stdin-thread error: %w", e.Err)
	}
	// note than stdin had panic or error other than io.EOF
	if r.isError != nil && !r.isError.Load() {
		r.isError.Store(true)
	}
	r.errorSink.AddError(err)
}

const (
	// buffer size of error channel from thread
	errChSize = 2
)
