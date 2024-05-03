/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"
	"sync/atomic"
)

type IdempotentCloser[C io.Closer] struct {
	closer  io.Closer
	isClose atomic.Bool
}

func NewIdemPotentCloser[C io.Closer](closer C) (idempotentCloser *IdempotentCloser[C]) {
	return &IdempotentCloser[C]{closer: closer}
}

func (c *IdempotentCloser[C]) Close() (err error) {
	if c.isClose.Load() || !c.isClose.CompareAndSwap(false, true) {
		return
	} else if c.closer == nil {
		return
	}
	err = c.closer.Close()
	return
}

func (c *IdempotentCloser[C]) IsClose() (isClose bool) { return c.isClose.Load() }
