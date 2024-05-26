/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"strconv"
	"sync"

	"github.com/haraldrudell/parl"
)

type CloseWait struct {
	closer      io.Closer
	label       string
	closeCount  parl.Atomic64[int]
	closeLock   sync.Mutex
	interceptor CloseInterceptor
	isClosed    parl.Awaitable
}

func NewCloseWait(closer any, label string, closeInterceptor CloseInterceptor) (closeWait *CloseWait) {
	parl.NilPanic("closeInterceptor", closeInterceptor)
	parl.NilPanic("stream", closer)
	if label == "" {
		label = "stream" + strconv.Itoa(streamNo.Add(1))
	}
	var closer2, _ = closer.(io.Closer)
	return &CloseWait{
		closer:      closer2,
		label:       label,
		interceptor: closeInterceptor,
	}
}

func (c *CloseWait) IsClosed() (isClosed bool) { return c.isClosed.IsClosed() }

func (c *CloseWait) Ch() { <-c.isClosed.Ch() }

func (c *CloseWait) Close() (err error) {
	defer parl.RecoverErr(func() parl.DA { return parl.A() }, &err)
	defer c.closeLock.Unlock()
	c.closeLock.Lock()

	return c.interceptor.CloseIntercept(c.closer, c.label, c.closeCount.Add(1))
}

var streamNo parl.Atomic64[int]
