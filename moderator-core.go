/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"sync"
)

const (
	defaultParallelism = 20
)

/*
ModeratorCore invokes functions at a limited level of parallelism.
ModeratorCore is a ticketing system.
ModeratorCore does not have a cancel feature.
 m := NewModeratorCore(20, ctx)
 m.Do(func() (err error) { // waiting here for a ticket
   // got a ticket!
   …
   return or panic // ticket automatically returned
 m.String() → waiting: 2(20)
*/
type ModeratorCore struct {
	parallelism uint64
	cond        *sync.Cond
	active      uint64 // behind lock
	waiting     uint64 // behind lock
}

// moderatorCore is a parl-private version of ModeratorCore
type moderatorCore struct {
	*ModeratorCore
}

// NewModerator creates a new Moderator used to limit parallelism
func NewModeratorCore(parallelism uint64) (mo *ModeratorCore) {
	if parallelism < 1 {
		parallelism = defaultParallelism
	}
	return &ModeratorCore{
		parallelism: parallelism,
		cond:        sync.NewCond(&sync.Mutex{}),
	}
}

// Do calls fn limited by the moderator’s parallelism.
// Do blocks until a ticket is available
// Do uses the same thread.
func (mo *ModeratorCore) Do(fn func()) {
	mo.getTicket()          // blocking
	defer mo.returnTicket() // we will always get a ticket, and it should be returned

	if fn != nil {
		fn()
	}
}

func (mo *ModeratorCore) getTicket() {
	mo.cond.L.Lock()
	defer mo.cond.L.Unlock()

	isWaiting := false
	for {
		if mo.active < mo.parallelism {
			mo.active++

			// maintain waiting counter
			if isWaiting {
				mo.waiting--
			}
			return
		}

		// maintain waiting counter
		if !isWaiting {
			isWaiting = true
			mo.waiting++
		}

		// block until cond.Notify or cond.Broadcast
		mo.cond.Wait()
	}
}

func (mo *ModeratorCore) returnTicket() {
	mo.cond.L.Lock()
	defer mo.cond.Signal()
	defer mo.cond.L.Unlock()

	mo.active--
}

func (mo *ModeratorCore) Status() (parallelism uint64, active uint64, waiting uint64) {
	parallelism = mo.parallelism
	mo.cond.L.Lock()
	defer mo.cond.L.Unlock()

	active = mo.active
	waiting = mo.waiting
	return
}

func (mo *ModeratorCore) String() (s string) {
	parallelism, active, waiting := mo.Status()
	if active < parallelism {
		s = fmt.Sprintf("available: %d(%d)", parallelism-active, parallelism)
	} else {
		s = fmt.Sprintf("waiting: %d(%d)", waiting, parallelism)
	}
	return
}
