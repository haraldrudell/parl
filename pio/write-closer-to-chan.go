/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type WriteCloserToChan struct {
	ch parl.NBChan[[]byte]
}

func NewWriteCloserToChan() (writeCloser io.WriteCloser) {
	return &WriteCloserToChan{}
}

func InitWriteCloserToChan(wcp *WriteCloserToChan) {}

func (wc *WriteCloserToChan) Write(p []byte) (n int, err error) {
	if wc.ch.DidClose() {
		err = perrors.ErrorfPF(ErrFileAlreadyClosed.Error())
		return
	}
	wc.ch.Send(p)
	n = len(p)
	return
}
func (wc *WriteCloserToChan) Close() (err error) {
	wc.ch.Close()
	return
}

func (wc *WriteCloserToChan) Ch() (readCh <-chan []byte) {
	return wc.ch.Ch()
}
