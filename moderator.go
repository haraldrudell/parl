/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"fmt"
	"sync"
)

const (
	defaultParallelism = 20
)

/*
Moderator invokes functions at a limited level of parallelism.
It is a ticketing system
 m := NewModerator(20, ctx)
 m.Do(func() (err error) { // waiting here for a ticket
   // got a ticket!
   …
   return or panic // ticket automatically returned
 m.String() → waiting: 2(20)
*/
type Moderator struct {
	parallelism uint64
	cond        *sync.Cond
	ctx         context.Context
	active      uint64 // behind lock
	waiting     uint64 // behind lock
}

// NewModerator creates a new Moderator used to limit parallelism
func NewModerator(parallelism uint64, ctx context.Context) (mo *Moderator) {
	if parallelism == 0 {
		parallelism = defaultParallelism
	}
	m := Moderator{parallelism: parallelism, ctx: ctx}
	m.cond = sync.NewCond(&sync.Mutex{})
	go m.shutdownThread()
	return &m
}

// Do calls fn limited by the moderator’s parallelism.
// If the moderator is shut down, ErrModeratorShutdown is returned
func (mo *Moderator) Do(fn func() error) (err error) {
	if fn == nil {
		return New("Moderator.Do with nil function")
	}
	if err = mo.getTicket(); err != nil {
		return
	}
	defer mo.returnTicket() // we will always get a ticket, and it should be returned
	return fn()
}

func (mo *Moderator) getTicket() (err error) {
	mo.cond.L.Lock()
	defer mo.cond.L.Unlock()
	isWaiting := false
	for {
		if mo.ctx.Err() != nil {
			err = mo.ctx.Err()
			break
		}
		if mo.active < mo.parallelism {
			mo.active++
			break
		}
		if !isWaiting {
			isWaiting = true
			mo.waiting++
		}
		mo.cond.Wait()
	}
	if isWaiting {
		mo.waiting--
	}
	return
}

func (mo *Moderator) returnTicket() {
	mo.cond.L.Lock()
	mo.cond.Signal()
	defer mo.cond.L.Unlock()
	mo.active--
}

func (mo *Moderator) shutdownThread() {
	<-mo.ctx.Done()
	mo.cond.Broadcast()
}

func (mo *Moderator) Status() (parallelism uint64, active uint64, waiting uint64, isShutdown bool) {
	parallelism = mo.parallelism
	mo.cond.L.Lock()
	active = mo.active
	waiting = mo.waiting
	mo.cond.L.Unlock()
	isShutdown = mo.ctx.Err() != nil
	return
}

func (mo *Moderator) String() (s string) {
	parallelism, active, waiting, sd := mo.Status()
	if active < parallelism {
		s = fmt.Sprintf("available: %d(%d)", parallelism-active, parallelism)
	} else {
		s = fmt.Sprintf("waiting: %d(%d)", waiting, parallelism)
	}
	if sd {
		s += " shutdown"
	}
	return
}
