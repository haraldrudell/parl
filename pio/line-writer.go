/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// LineReader reads a stream one line per Read invocation.
package pio

import (
	"io"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

const (
	newLineWriter = byte('\n')
)

type LineWriter struct {
	writeCloser io.WriteCloser
	filter      func(line *[]byte, isLastLine bool) (skipLine bool, err error)

	dataLock sync.Mutex
	isClosed bool
	data     []byte
}

var _ io.WriteCloser = &LineWriter{}

func NewLineWriter(writeCloser io.WriteCloser,
	filter func(line *[]byte, isLastLine bool) (skipLine bool, err error)) (lineWriter *LineWriter) {
	if writeCloser == nil {
		panic(perrors.NewPF("writeCloser cannot be nil"))
	}
	return &LineWriter{writeCloser: writeCloser, filter: filter}
}

// Write saves data in slice and returns all bytes written or ErrFileAlreadyClosed
func (wc *LineWriter) Write(p []byte) (n int, err error) {
	wc.dataLock.Lock()
	defer wc.dataLock.Unlock()

	if wc.isClosed {
		err = perrors.ErrorfPF("%w", ErrFileAlreadyClosed)
		return // closed return
	}

	// consume data
	length := len(p)
	for n < length {
		index := slices.Index(p[n:], newLineWriter)

		// check for p ending without newline
		if index == -1 {
			wc.data = append(wc.data, p[n:]...) // save in buffer
			n = length                          // pretend data was written
			break
		}

		index += n + 1 // include newline, make index in p
		wc.data = append(wc.data, p[n:index]...)
		if err = wc.processLine(); err != nil {
			return
		}
		wc.data = wc.data[:0]
		n = index
	}

	return // good write return
}

func (wc *LineWriter) processLine() (err error) {

	// apply filter
	if wc.filter != nil {
		var skipLine bool
		if parl.RecoverInvocationPanic(func() {
			skipLine, err = wc.filter(&wc.data, wc.isClosed)
		}, &err); err != nil || skipLine {
			return
		}
	}

	// write line to writeCloser
	length := len(wc.data)
	var n int
	for n < length {
		var n0 int
		if n0, err = wc.writeCloser.Write(wc.data[n:]); err != nil {
			return
		}
		n += n0
	}

	return
}

// Close closes
func (wc *LineWriter) Close() (err error) {
	wc.dataLock.Lock()
	defer wc.dataLock.Unlock()

	if wc.isClosed {
		err = perrors.ErrorfPF("%w", ErrFileAlreadyClosed)
		return // closed return
	}

	wc.isClosed = true
	if len(wc.data) > 0 {
		err = wc.processLine()
	}

	return
}
