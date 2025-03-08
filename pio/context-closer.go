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

// ContextCloser is an idempotent io.Closer
//   - implements [io.Closer]
type ContextCloser struct {
	closer   io.Closer
	isClosed atomic.Bool
}

// NewContextCloser returns a an idempotent [io.Closer]
//   - closer may be nil
//   - panic-free idempotent observable
func NewContextCloser(closer io.Closer, fieldp ...*ContextCloser) (contextCloser *ContextCloser) {

	if len(fieldp) > 0 {
		contextCloser = fieldp[0]
	}
	if contextCloser == nil {
		contextCloser = &ContextCloser{}
	}

	*contextCloser = ContextCloser{closer: closer}
	return
}

// Close closes the io.Closer
//   - if Close has already been invoked, noop, no error
//   - if io.Closer is nil, noop, no error
//   - panic-free idempotent
func (c *ContextCloser) Close() (err error) {
	if c.isClosed.Load() {
		return // Close already invoked
	} else if !c.isClosed.CompareAndSwap(false, true) {
		return // another thread already closed
	} else if c.closer == nil {
		return // closer is nil
	}
	parl.Close(c.closer, &err)
	return
}

// IsCloseable indicates whether an [io.Closer] is present that can be closed
func (c *ContextCloser) IsCloseable() (isCloseable bool) {
	return c.closer != nil
}
