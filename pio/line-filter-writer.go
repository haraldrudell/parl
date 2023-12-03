/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"io/fs"
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
)

const (
	newLineWriter = byte('\n')
)

// LineFilterFunc receives lines as they are written to the writer
//   - can modify the line
//   - can skip the line using skipLine
//   - can return error
type LineFilterFunc func(line *[]byte, isLastLine bool) (skipLine bool, err error)

// LineFilterWriter is a writer that filters each line using a filter function
type LineFilterWriter struct {
	writeCloser io.WriteCloser
	filter      LineFilterFunc

	dataLock sync.Mutex
	isClosed bool
	data     []byte
}

var _ io.WriteCloser = &LineFilterWriter{}

// NewLineFilterWriter is a writer that filters each line using a filter function
func NewLineFilterWriter(writeCloser io.WriteCloser, filter LineFilterFunc) (lineWriter *LineFilterWriter) {
	if writeCloser == nil {
		panic(parl.NilError("writeCloser"))
	} else if filter == nil {
		panic(parl.NilError("filter"))
	}
	return &LineFilterWriter{writeCloser: writeCloser, filter: filter}
}

// Write saves data in slice and returns all bytes written or ErrFileAlreadyClosed
func (wc *LineFilterWriter) Write(p []byte) (n int, err error) {
	wc.dataLock.Lock()
	defer wc.dataLock.Unlock()

	if wc.isClosed {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
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

func (w *LineFilterWriter) processLine() (err error) {

	// apply filter
	if w.filter != nil {
		var skipLine bool
		if skipLine, err = w.invokeFilter(); err != nil || skipLine {
			return
		}
	}

	// write line to writeCloser
	length := len(w.data)
	var n int
	for n < length {
		var n0 int
		if n0, err = w.writeCloser.Write(w.data[n:]); err != nil {
			return
		}
		n += n0
	}

	return
}

// Close closes
func (wc *LineFilterWriter) Close() (err error) {
	wc.dataLock.Lock()
	defer wc.dataLock.Unlock()

	if wc.isClosed {
		err = perrors.ErrorfPF("%w", fs.ErrClosed)
		return // closed return
	}

	wc.isClosed = true
	if len(wc.data) > 0 {
		err = wc.processLine()
	}

	return
}

// invokeFilter captures a panic in the filter function
func (w *LineFilterWriter) invokeFilter() (skipLine bool, err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	skipLine, err = w.filter(&w.data, w.isClosed)

	return
}
