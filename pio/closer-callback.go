/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// Go does not allow promotion of a type parameter
// https://github.com/golang/go/issues/49030
// interfaces:
var _ io.ReadCloser
var _ io.ReadSeekCloser
var _ io.ReadWriteCloser
var _ io.WriteCloser
var _ io.Closer

// CloserCallbacker implements a close callback for io.Closer
type CloserCallbacker struct {
	closeCallback func(err error) (e error)
	isClosed      atomic.Bool
	wg            sync.WaitGroup
}

func (cc *CloserCallbacker) Close(closer io.Closer) (err error) {
	parl.RecoverInvocationPanic(func() {
		err = closer.Close()
	}, &err)
	if cc.isClosed.CompareAndSwap(false, true) {
		if cc.closeCallback != nil {
			var e error
			parl.RecoverInvocationPanic(func() {
				e = cc.closeCallback(err)
			}, &e)
			err = perrors.AppendError(err, e)
		}
		cc.wg.Done()
	}
	return
}

func (cc *CloserCallbacker) IsClosed() (isClosed bool) {
	return cc.isClosed.Load()
}

func (cc *CloserCallbacker) Wait() {
	cc.wg.Wait()
}

type WriteCloserCallbacker struct {
	io.WriteCloser
	CloserCallbacker
}

func NewWriteCloserCallbacker(closeCallback func(err error) (e error), writeCloser io.WriteCloser) (writeCloserCallbacker io.WriteCloser) {
	cc := WriteCloserCallbacker{
		WriteCloser:      writeCloser,
		CloserCallbacker: CloserCallbacker{closeCallback: closeCallback},
	}
	cc.CloserCallbacker.wg.Add(1)
	return &cc
}

func (cc *WriteCloserCallbacker) Close() (err error) {
	return cc.CloserCallbacker.Close(cc.WriteCloser)
}

type ReadCloserCallbacker struct {
	io.ReadCloser
	CloserCallbacker
}

func NewReadCloserCallbacker(closeCallback func(err error) (e error), readCloser io.ReadCloser) (readCloserCallbacker io.ReadCloser) {
	cc := ReadCloserCallbacker{
		ReadCloser:       readCloser,
		CloserCallbacker: CloserCallbacker{closeCallback: closeCallback},
	}
	cc.CloserCallbacker.wg.Add(1)
	return &cc
}

func (cc *ReadCloserCallbacker) Close() (err error) {
	return cc.CloserCallbacker.Close(cc.ReadCloser)
}
