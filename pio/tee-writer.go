/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// TeeWriter is a writer that copies its writes to one or more other writers.
package pio

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// TeeWriter is a writer that copies its writes to one or more other writers.
type TeeWriter struct {
	closeCallback func() (err error)
	writers       []io.Writer
	isClosed      atomic.Bool
	wg            sync.WaitGroup
}

// TeeWriter is a writer that copies its writes to one or more other writers.
func NewTeeWriter(closeCallback func() (err error), writers ...io.Writer) (teeWriter io.WriteCloser) {
	length := len(writers)
	if length == 0 {
		panic(perrors.NewPF("Must have one or more writers, writers is empty"))
	}
	t := TeeWriter{closeCallback: closeCallback, writers: make([]io.Writer, length)}
	for i, w := range writers {
		if w == nil {
			panic(perrors.ErrorfPF("Writers#%d nil", i))
		}
		t.writers[i] = w
	}
	t.wg.Add(1)
	return &t
}

func (tw *TeeWriter) Write(p []byte) (n int, err error) {
	if tw.isClosed.Load() {
		err = perrors.NewPF("Write after Close")
		return
	}
	length := len(p)
	for _, writer := range tw.writers {
		written := 0
		for written < length {
			n, err = writer.Write(p)
			written += n
			if err != nil {
				return // write error return
			}
		}
	}
	return // good write return
}

func (w *TeeWriter) Close() (err error) {

	// prevent multiple Close invocations
	if !w.isClosed.CompareAndSwap(false, true) {
		err = perrors.NewPF("Second Close invocation")
		return
	}

	// invoke callback if there is one
	if w.closeCallback != nil {
		err = w.invokeCallback()
	}

	return
}

func (w *TeeWriter) invokeCallback() (err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	err = w.closeCallback()

	return
}
