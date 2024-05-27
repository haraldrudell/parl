/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"io/fs"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type WriteCloserToChan struct{ ch parl.AwaitableSlice[[]byte] }

func NewWriteCloserToChan() (writeCloser io.WriteCloser) {
	return &WriteCloserToChan{}
}

func InitWriteCloserToChan(wcp *WriteCloserToChan) {}

func (wc *WriteCloserToChan) Write(p []byte) (n int, err error) {
	if parl.IsClosed[string](&wc.ch) {
		err = perrors.ErrorfPF(fs.ErrClosed.Error())
		return
	}
	wc.ch.Send(p)
	n = len(p)
	return
}
func (wc *WriteCloserToChan) Close() (err error) {
	wc.ch.EmptyCh()
	return
}

func (wc *WriteCloserToChan) Ch() (readCh parl.ClosableAllSource[[]byte]) { return &wc.ch }
