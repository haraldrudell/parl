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
		result *ReadResult
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
	<-result.IsReady
	r.Close()
	<-result.Ch
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
	<-result.IsReady
	// wait 1 ms
	<-time.NewTimer(time.Millisecond).C
	if result.IsExit.Load() {
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
	<-result.Ch
	if result.err != nil {
		t.Errorf("Read err %s", perrors.Short(err))
	}
	if !slices.Equal(result.p, slice1) {
		t.Errorf("Read %v exp %v", result.p, slice1)
	}
}

type ReadResult struct {
	Ch      chan *ReadResult
	IsReady chan struct{}
	IsExit  atomic.Bool
	p       []byte
	err     error
}

func newReadResult(n int) (result *ReadResult) {
	return &ReadResult{
		Ch:      make(chan *ReadResult, 1),
		IsReady: make(chan struct{}),
		p:       make([]byte, n),
	}
}

func (r *ReadResult) read(reader io.Reader) {
	defer func() { r.Ch <- r }()
	defer r.IsExit.Store(true)

	close(r.IsReady)
	var n int
	n, r.err = reader.Read(r.p)
	r.p = r.p[:n]
}
