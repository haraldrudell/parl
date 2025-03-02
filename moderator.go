/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

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
	moderatorCore
	//ctx context.Context
}

// moderatorCore is a parl-private version of ModeratorCore
type moderatorCore struct {
	*ModeratorCore
}

// NewModerator creates a cancelable Moderator used to limit parallelism
// func NewModerator(parallelism uint64, ctx context.Context) (mo *Moderator) {
// 	return NewModeratorFromCore(NewModeratorCore(parallelism), ctx)
// }

// NewModeratorFromCore allows multiple cancelable Moderators to share
// a ModeratorCore ticket pool
// func NewModeratorFromCore(mc *ModeratorCore, ctx context.Context) (mo *Moderator) {
// 	if ctx == nil {
// 		panic(perrors.New("NewModerator with nil context"))
// 	}
// 	m := Moderator{
// 		moderatorCore: moderatorCore{mc},
// 		ctx:           ctx,
// 	}
// 	go m.shutdownThread()
// 	return &m
// }

// func (mo *Moderator) DoErr(fn func() error) (err error) {
// 	if err = mo.getTicket(); err != nil {
// 		return // failed to obtain ticket due to cancelation
// 	}
// 	defer mo.returnTicket() // we will always get a ticket, and it should be returned

// 	if fn != nil {
// 		err = fn()
// 	}
// 	return
// }

// // Do calls fn limited by the moderator’s parallelism.
// // If the moderator is shut down, ErrModeratorShutdown is returned
// func (mo *Moderator) Do(fn func()) {
// 	if err := mo.getTicket(); err != nil {
// 		panic(err) // failed to obtain ticket due to cancelation
// 	}
// 	defer mo.returnTicket() // we will always get a ticket, and it should be returned

// 	if fn != nil {
// 		fn()
// 	}
// }

// func (mo *Moderator) getTicket() (err error) {
// 	mo.queue.L.Lock()
// 	defer mo.queue.L.Unlock()

// 	isWaiting := false
// 	for {

// 		// check for cancelation
// 		if mo.ctx.Err() != nil {
// 			err = mo.ctx.Err()
// 			return // moderator cancelled return
// 		}

// 		// check for available ticket
// 		if mo.active < mo.parallelism {
// 			mo.active++

// 			// maintain waiting counter
// 			if isWaiting {
// 				mo.waiting--
// 			}
// 			return // obtained ticket return
// 		}

// 		// maintain waiting counter
// 		if !isWaiting {
// 			isWaiting = true
// 			mo.waiting++
// 		}

// 		// block until cond.Notify or cond.Broadcast
// 		mo.queue.Wait()
// 	}
// }

// func (mo *Moderator) shutdownThread() {
// 	<-mo.ctx.Done()
// 	mo.queue.Broadcast()
// }

// func (mo *Moderator) Status() (parallelism uint64, active uint64, waiting uint64, isShutdown bool) {
// 	parallelism, active, waiting = mo.moderatorCore.Status()
// 	isShutdown = mo.ctx.Err() != nil
// 	return
// }

// func (mo *Moderator) String() (s string) {
// 	s = mo.moderatorCore.String()
// 	if mo.ctx.Err() != nil {
// 		s += " shutdown"
// 	}
// 	return
// }
