/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// ReadWriteCloserSlice is a read-writer with a slice as intermediate storage. thread-safe.
package pio

import (
	"errors"
	"io"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/haraldrudell/parl/perrors"
)

func TestReadWriteCloserSlice(t *testing.T) {
	var (
		slice1    = []byte{1}
		exp1      = 1
		bufLength = 128
	)

	var (
		n      int
		err    error
		p, pn  []byte
		result *readResult
	)

	// Read Write Close
	var r *ReadWriteCloserSlice
	var reset = func() {
		r = NewReadWriteCloserSlice()
	}

	// Write then Read should work
	reset()
	// Write should write 1 byte without error
	n, err = r.Write(slice1)
	if err != nil {
		t.Errorf("Write err %s", perrors.Short(err))
	}
	if n != exp1 {
		t.Errorf("Write n %d exp %d", n, exp1)
	}
	// Read should get the byte from data
	p = make([]byte, bufLength)
	n, err = r.Read(p)
	if err != nil {
		t.Errorf("Read err %s", perrors.Short(err))
	}
	pn = p[:n]
	if !slices.Equal(pn, slice1) {
		t.Errorf("Read %v exp %v", pn, slice1)
	}

	// Read ending from close should work
	reset()
	result = newReadResult(bufLength)
	go result.read(r)
	<-result.isReady
	r.Close()
	<-result.ch
	if !errors.Is(result.err, io.EOF) {
		t.Errorf("Read Close err not EOF: %s", perrors.Short(err))
	}
	if len(result.p) > 0 {
		t.Errorf("Read Close len %d exp 0", result.p)
	}

	// Read then Write should work
	reset()
	// Read should block
	result = newReadResult(bufLength)
	go result.read(r)
	<-result.isReady
	// wait 1 ms
	<-time.NewTimer(time.Millisecond).C
	if result.isExit.Load() {
		t.Error("Read completed with no data")
	}
	// Write should succeed
	n, err = r.Write(slice1)
	if err != nil {
		t.Errorf("Write err %s", perrors.Short(err))
	}
	if n != exp1 {
		t.Errorf("Write n %d exp %d", n, exp1)
	}
	// Read should have the data
	<-result.ch
	if result.err != nil {
		t.Errorf("Read err %s", perrors.Short(err))
	}
	if !slices.Equal(result.p, slice1) {
		t.Errorf("Read %v exp %v", result.p, slice1)
	}
}

// readResult is a reader with observable progress
type readResult struct {
	ch      chan *readResult
	isReady chan struct{}
	isExit  atomic.Bool
	p, p0   []byte
	err     error
}

// newReadResult returns a reader with inspectable progress
//   - n: read buffer size
//   - result.IsReady closes once the reading thread is ready
//   - result.Ch sends the result object once a Read completes
//   - — result.P and result.Err contains Read result
//   - result.IsExit is set to true on thread exit
func newReadResult(n int) (result *readResult) {
	return &readResult{
		ch:      make(chan *readResult, 1),
		isReady: make(chan struct{}),
		p0:      make([]byte, n),
	}
}

// read attempts a Read: can only be invoked once
//   - r.IsReady indicates thread is invoking Read
//   - r.ch inidcates Read concluded
//   - r.IsExit is atomic Bool for reead concluded
//   - r.p and r.,err contains Read outcome
func (r *readResult) read(reader io.Reader) {
	if r.isExit.Load() {
		panic("multiple invocations")
	}
	defer func() { r.ch <- r }()
	defer r.isExit.Store(true)

	close(r.isReady)
	var n int
	n, r.err = reader.Read(r.p0)
	r.p = r.p0[:n]
}

// awaitCh awaits a pending readResult.read with optional timeout
//   - isChClosed true: Read concluded
func (r *readResult) awaitCh(timeout time.Duration) (isChClosed bool) {
	var C <-chan time.Time
	if timeout > 0 {
		var timer = time.NewTimer(timeout)
		C = timer.C
		defer timer.Stop()
	}
	select {
	case <-r.ch:
		return true
	case <-C:
		return
	}
}
