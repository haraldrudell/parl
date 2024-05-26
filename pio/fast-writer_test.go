/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"bytes"
	"errors"
	"io"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

func TestFastWriter(t *testing.T) {
	const (
		// timeout when debugging
		//timeout = 0
		// timeout when testing
		timeout = time.Millisecond
	)
	var (
		byte1 = []byte{1}
	)

	var (
		buffer         []byte
		isTimeout      bool
		err, writerErr error
		await          *awaiter
	)

	// Buffer Close Length Write
	var writer *FastWriter
	var output *ThreadSafeRwc
	var reset = func() {
		output = NewThreadSafeRwc().(*ThreadSafeRwc)
		writer = NewFastWriter(output).(*FastWriter)
	}

	// Write invocation should result in write to output
	reset()
	// Write commits result to buffer which is then read by the FastWriter thread
	writer.Write(byte1)
	// await byte, timeout or error:
	//	- writer.Write storing byte in buffer and alerting thread
	//	- thread writing the byte to output
	//	- readByte reading the byte from output
	await = newAwaiter(output, timeout)
	buffer, isTimeout, err = await.readByte()
	writerErr, _ = writer.err.Error()
	t.Logf(
		"output Read: p: %v err: %s isTimeout: %t"+
			"\nFastWriter buffer: %v thread-running: %t err: %s isClose: %t",
		buffer, perrors.Short(err), isTimeout,
		writer.Buffer(), !writer.threadAwait.IsClosed(), perrors.Short(writerErr), writer.closeOnce.IsInvoked(),
	)
	if isTimeout {
		// the thread is unrecoverably blocked in Read, so terminate the test
		t.Fatal("FAIL Read timeout")
	}
	if err != nil {
		t.Errorf("FAIL Read err %s", perrors.Short(err))
	}
	if !slices.Equal(buffer, byte1) {
		t.Errorf("FAIL bad Write %v exp %v", buffer, byte1)
	}
}

func TestFastWriterThread(t *testing.T) {

	var (
		err error
	)

	var output = NewThreadSafeRwc()
	var writer = NewFastWriter(output).(*FastWriter)

	// on Close, the thread should exit
	err = writer.Close()
	if err != nil {
		t.Errorf("Close err %s", perrors.Short(err))
	}
	if !writer.threadAwait.IsClosed() {
		t.Error("no thread did exit")
	}
}

func TestFastWriterToClosed(t *testing.T) {

	var (
		err  error
		byt0 []byte
		n    int
	)

	var output = bytes.NewBuffer(make([]byte, 0, 1))
	var writer = NewFastWriter(output).(*FastWriter)

	// Write after Close should error
	err = writer.Close()
	if err != nil {
		t.Errorf("Close err %s", perrors.Short(err))
	}
	n, err = writer.Write(byt0)
	if n != 0 {
		t.Errorf("Write n %d exp 0", n)
	}
	if !errors.Is(err, os.ErrClosed) {
		t.Errorf("Write after Close not ErrClosed %s", perrors.Short(err))
	}

	// Close after Close, no write error
	err = writer.Close()
	if err != nil {
		t.Errorf("Close Close err %s", perrors.Short(err))
	}
}

// awaiter reads one byte from a stream with timeout
type awaiter struct {
	// the reader the byte is read from
	reader io.Reader
	// optional timeout, 0 for none
	timeout time.Duration
}

// newAwaiter returns an object that reads one byte from a stream with timeout
func newAwaiter(reader io.Reader, timeout time.Duration) (a *awaiter) {
	return &awaiter{
		reader:  reader,
		timeout: timeout,
	}
}

// readByte attempts to read a byte from a.reader in a separate thread
//   - blocks until timeout or byte read
//   - isTimeout false: byt is the read byte, err is any error from Read
//   - isTimeout true: Read invocation timed out. byt and err are nil
//   - readByte can be invoked multiple times
//   - any timeout leaves a thread blocked in [io.Reader.Read]
func (a *awaiter) readByte() (byt []byte, isTimeout bool, err error) {

	// create thread reading from reader
	var o = newOutcome(a.reader, &byt, &err)
	go o.readByteThread()

	// configure possible timeout
	var C <-chan time.Time
	if a.timeout > 0 {
		var timer = time.NewTimer(a.timeout)
		defer timer.Stop()
		C = timer.C
	}

	// await outcome
	select {
	case <-o.ch:
	case <-C:
		o.setIsTimeout()
	}
	return
}

// outcome contains one instance of a read outcome
type outcome struct {
	// the reader to read from
	reader io.Reader
	// where to store the read byte
	byt *[]byte
	// where to store any error
	errp *error
	// ch makes thread awaitable
	ch chan struct{}
	// makes isTimeout isDone thread-safe
	timeoutLock sync.Mutex
	// true if the result is timeout
	isTimeout bool
	// true if a byte was successfully read
	isDone bool
}

// newOutcome returns an outcome either Read result from readByteThread or timeout
func newOutcome(reader io.Reader, byt *[]byte, errp *error) (o *outcome) {
	return &outcome{
		reader: reader,
		byt:    byt,
		errp:   errp,
		ch:     make(chan struct{}),
	}
}

// readByteThread reads indefinitely
func (o *outcome) readByteThread() {
	defer close(o.ch)

	// invoke Read
	var b = []byte{0}
	var n, err = o.reader.Read(b)

	// update result
	o.timeoutLock.Lock()
	defer o.timeoutLock.Unlock()

	// is it timeout?
	if o.isTimeout {
		return
	}

	// update values
	o.isDone = true
	*o.byt = b[:n]
	*o.errp = err
}

// setIsTimeout sets end-state timeout if isDone not already set
func (o *outcome) setIsTimeout() {
	o.timeoutLock.Lock()
	defer o.timeoutLock.Unlock()

	// attempt to set isTimeout true
	if o.isDone {
		return // already done
	}
	o.isTimeout = true
}
