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

// it is desirable to modify the Close method for any interface or type
// having a Close method
//   - Go does not allow promotion of a type parameter
//     https://github.com/golang/go/issues/49030
//   - this limitation means a generic Close wrapper can not be defined
//     that would promote whatever methods the generic type provides
//     in addition to Close
//   - the types providing a Close method that it would be desirable to promote:
var _ io.ReadCloser
var _ io.ReadSeekCloser
var _ io.ReadWriteCloser
var _ io.WriteCloser
var _ io.Closer

type CloseCallback func(err error) (e error)

// CloseCallbacker implements a close callback for io.Closer
// 231128 unused
type CloseCallbacker struct {
	closer        io.Closer
	closeCallback CloseCallback
	isClosed      parl.Awaitable
}

var _ io.Closer = &CloseCallbacker{}

func NewCloseCallbacker(closer io.Closer, closeCallback CloseCallback) (closeCallbacker *CloseCallbacker) {
	if closer == nil {
		panic(parl.NilError("closer"))
	} else if closeCallback == nil {
		panic(parl.NilError("closeCallback"))
	}
	return &CloseCallbacker{
		closer:        closer,
		closeCallback: closeCallback,
	}
}

func (c *CloseCallbacker) Close() (err error) {
	if !c.isClosed.Close() {
		return // already closed
	}
	parl.Close(c.closer, &err)
	var e = c.invokeCloseCallback(err)
	if e != nil && e != err {
		err = perrors.AppendError(err, e)
	}
	return
}

func (c *CloseCallbacker) IsClosed() (isClosed bool) { return c.isClosed.IsClosed() }

func (c *CloseCallbacker) Wait() { <-c.isClosed.Ch() }

func (c *CloseCallbacker) invokeCloseCallback(e error) (err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)

	err = c.closeCallback(e)

	return
}
