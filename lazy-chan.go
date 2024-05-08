/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import "sync/atomic"

type LacyChan[T any] struct{ cp atomic.Pointer[chan T] }

func (c *LacyChan[T]) Get(n ...int) (ch chan T) {
	if chp := c.cp.Load(); chp != nil {
		ch = *chp
		return
	}
	var n0 int
	if len(n) > 0 && n[0] >= 0 {
		n0 = n[0]
	}
	ch = make(chan T, n0)
	if c.cp.CompareAndSwap(nil, &ch) {
		return
	}
	ch = *c.cp.Load()
	return
}
