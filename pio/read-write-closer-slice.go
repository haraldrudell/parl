/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"sync"

	"github.com/haraldrudell/parl/perrors"
)

type ReadWriteCloserSlice struct {
	lock     sync.Mutex
	isClosed bool
	data     []byte
}

func NewReadWriteCloserSlice() (readWriteCloser io.ReadWriteCloser) {
	return &ReadWriteCloserSlice{}
}

func InitReadWriteCloserSlice(wcp *ReadWriteCloserSlice) {}

func (wc *ReadWriteCloserSlice) Write(p []byte) (n int, err error) {
	wc.lock.Lock()
	defer wc.lock.Unlock()

	if wc.isClosed {
		err = perrors.ErrorfPF("%w", ErrFileAlreadyClosed)
		return // closed return
	}

	// consume data
	wc.data = append(wc.data, p...)
	n = len(p)

	return // good write return
}

func (wc *ReadWriteCloserSlice) Read(p []byte) (n int, err error) {
	wc.lock.Lock()
	defer wc.lock.Unlock()

	data := wc.data
	d := len(data)

	// EOF
	if wc.isClosed && d == 0 {
		err = io.EOF
		return // end of file return
	}

	// copy data
	copy(p, data)

	// all data consumed
	n = len(p) // assume p is shorter
	if d <= n {
		n = d              // n is length of the shorter data bytes
		wc.data = data[:0] // all data was consumed
		return             // all data submitted return
	}

	// p was shorter than data
	// n already has the shorter len(p) value
	wc.data = data[n:] // remove consumed bytes from data
	return             // p filled return
}

func (wc *ReadWriteCloserSlice) Close() (err error) { return }
