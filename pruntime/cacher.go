/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"sync"
	"sync/atomic"
)

// CacheMechanic ensures that an init function is executed exactly once
//   - sync.Once using a known number of stack frames
//   - initialization-free
type CacheMechanic struct {
	// first access makes threads wait until data is available
	//	- subsequent accesses is atomic performance
	initLock sync.Mutex
	// written inside lock
	//	- isReady read provides happens-before
	isReady atomic.Bool
}

// EnsureInit ensures data is loaded exactly once
//   - initFunc loads data
//   - invocations after first EnsureInit return are atomic performance
//   - first invocation is locked performance
//   - subsequent invocations prior to first EnsureInit return are held waiting
//   - —
//   - upon return, it is guaranteed that init has completed
//   - order of thread-returns is not guaranteed
func (c *CacheMechanic) EnsureInit(initFunc func()) {

	// ouside lock fast check
	if c.isReady.Load() {
		return // already initialized
	}

	// first thread will win, other threads will wait
	c.initLock.Lock()
	defer c.initLock.Unlock()

	// inside lock check
	if c.isReady.Load() {
		return // already initialized by other thread
	}

	// winner thread does init
	initFunc()

	// flag values available
	c.isReady.Store(true)
}
