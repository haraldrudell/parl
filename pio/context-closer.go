/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pio

import (
	"io"
	"sync/atomic"

	"github.com/haraldrudell/parl"
)

type ContextCloser struct {
	closer   io.Closer
	isClosed atomic.Bool
}

func NewContextCloser(closer io.Closer) (contextCloser *ContextCloser) {
	return &ContextCloser{closer: closer}
}

func (c *ContextCloser) Close() (err error) {
	if !c.isClosed.CompareAndSwap(false, true) {
		return
	} else if c.closer == nil {
		return
	}
	parl.Close(c.closer, &err)
	return
}

func (c *ContextCloser) IsCloseable() (isCloseable bool) {
	return c.closer != nil
}
