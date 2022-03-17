/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	defaultParallelism = 20
	uint64MinusOne     = ^uint64(0)
)

// Moderator invokes functions at a limited level of parallelism
type Moderator struct {
	parallelism  uint64
	shutdownLock sync.Once
	isShutdown   AtomicBool
	dataLock     sync.Mutex
	available    uint64 // behind dataLock
	waitCount    uint64 // atomic
	waitLock     sync.Mutex
}

var ErrModeratorShutdown = errors.New("Moderator shut down")

// NewModerator creates a new Moderator used to limit parallelism
func NewModerator(parallelism uint64) (mo *Moderator) {
	if parallelism == 0 {
		parallelism = defaultParallelism
	}
	return &Moderator{parallelism: parallelism, available: parallelism}
}

// Do calls fn limited by the moderator’s parallelism.
// If the moderator is shut down, ErrModeratorShutdown is returned
func (mo *Moderator) Do(fn func() error) (err error) {
	if fn == nil {
		panic(New("Moderator.Do with nil function"))
	}
	if mo.isShutdown.IsTrue() {
		return ErrModeratorShutdown
	}
	defer mo.returnTicket() // we will always get a ticket, and it should be returned
	if err = mo.getTicket(); err != nil {
		return // shutdown
	}
	return fn()
}

func (mo *Moderator) getTicket() (err error) {
	if mo.getAvailableTicket() {
		return // a ticket was available
	}

	// wait for ticket
	mo.waitLock.Lock()

	// we now have a ticket!
	if mo.isShutdown.IsTrue() {
		return ErrModeratorShutdown
	}
	return
}

func (mo *Moderator) getAvailableTicket() (hasTicket bool) {
	mo.dataLock.Lock()
	defer mo.dataLock.Unlock()
	if mo.available > 0 {
		mo.available--
		if mo.available == 0 {
			mo.waitLock.Lock() // the next thread will wait
		}
		return true
	}
	mo.waitCount++
	return
}

func (mo *Moderator) returnTicket() {
	mo.dataLock.Lock()
	defer mo.dataLock.Unlock()

	// hand ticket to the queue
	if mo.waitCount > 0 {
		mo.waitCount--
		mo.waitLock.Unlock()
		return
	}
	if mo.available == 0 {
		mo.waitLock.Unlock()
	}

	// note an additional ticket available
	mo.available++
}

func (mo *Moderator) Status() (parallelism uint64, available uint64, waiting uint64, isShutdown bool) {
	parallelism = mo.parallelism
	mo.dataLock.Lock()
	available = mo.available
	mo.dataLock.Unlock()
	waiting = atomic.LoadUint64(&mo.waitCount)
	isShutdown = mo.isShutdown.IsTrue()
	return
}

func (mo *Moderator) String() (s string) {
	p, a, w, sd := mo.Status()
	if a > 0 {
		s = fmt.Sprintf("available: %d(%d)", a, p)
	} else {
		s = fmt.Sprintf("waiting: %d(%d)", w, p)
	}
	if sd {
		s += " shutdown"
	}
	return
}

func (mo *Moderator) shutdown() {
	mo.isShutdown.Set()
}

func (mo *Moderator) Shutdown() {
	mo.shutdownLock.Do(mo.shutdown)
}
