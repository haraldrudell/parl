/*
© 2024–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"io"
	"sync"
	"sync/atomic"
)

type IdempotentCloser[C io.Closer] struct {
	closer  io.Closer
	isClose atomic.Bool
	wg      sync.WaitGroup
}

// NewIdemPotentCloser
//   - C: type parameter for any type implementing [io.Closer]
//   - closer: the value being closed only once
//   - fieldp: optional fieldpointer
func NewIdemPotentCloser[C io.Closer](closer C, fieldp ...*IdempotentCloser[C]) (idempotentCloser *IdempotentCloser[C]) {

	// get pointer
	if len(fieldp) > 0 {
		idempotentCloser = fieldp[0]
	}
	if idempotentCloser == nil {
		idempotentCloser = &IdempotentCloser[C]{}
	}

	idempotentCloser.closer = closer
	idempotentCloser.wg.Add(1)

	return
}

func (c *IdempotentCloser[C]) Close() (err error) {

	// pick winner
	if c.isClose.Load() || !c.isClose.CompareAndSwap(false, true) {

		// losers wait
		c.wg.Wait()
		return
	}
	defer c.wg.Done()

	// winner closes
	if c.closer == nil {
		return
	}
	err = c.closer.Close()

	return
}

func (c *IdempotentCloser[C]) IsClose() (isClose bool) { return c.isClose.Load() }
