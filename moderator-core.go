/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"

	"github.com/haraldrudell/parl/perrors"
)

var (
	// [NewModeratorCorep]
	NoModeratorCore *ModeratorCore
)

// ModeratorCore limits parallelism
//   - ModeratorCore is a ticketing system
//   - ModeratorCore does not have a cancel feature
//   - during low contention: atomic performance
//   - during high-contention: lock performance
type ModeratorCore struct {
	// parallelism is the maximum number of tickets issued > 0
	parallelism int
	// active is number of currently issued tickets
	//	- less than parallelism: tickets are obtained via atomics.
	//		Moderator in atomic mode.
	//	- equal to parallelism: tickets are awaited and returned via lock.
	//		Moderator in lock-state.
	active Atomic64[int]
	// number of ticket-seeking threads attempting to enter critical section
	//	- incremented prior to entering critical section
	//	- decremented inside critical section
	seekers Atomic64[int]
	// lock forms critical section while in moderator lock-state
	//	- parl.[Mutex] one-liner
	lock Mutex
	// didFirst is false if ticketQueue has not been locked
	// 	- behind lock
	didFirst bool
	// heldTickets holds tickets for threads that did increment seekers
	// but have yet to enter critical section
	//	- behind lock
	heldTickets int
	// number of threads that are awaiting ticket from ticketQueue
	//	- written behind lock
	waiting Atomic64[int]
	// ticketQueue forming ticket queue when moderator in lock-state:
	//	- orderly first-come-first-served
	//	- parl.[Mutex] one-liner
	ticketQueue Mutex
}

// NewModerator returns a parallelism limiter
//   - parallelism > 0: the number of concurrent tickets issued
//   - parallelism < 1: defaultnumber of tickets: 20
//   - —
//   - [ModeratorCore.Ticket] awaits available ticket
//   - cancel is separate from ticketing.
//     A ticket-holder detecting cancel simply returns the ticket
//
// Usage:
//
//	var m = NewModeratorCore(20)
//	…
//	 // blocks here for available ticket
//	defer m.Ticket().ReturnTicket()
//	// holds ticket here until return or panic
func NewModeratorCore(parallelism int) (m *ModeratorCore) {
	return NewModeratorCorep(NoModeratorCore, parallelism)
}

// NewModeratorp returns a parallelism limiter
func NewModeratorCorep(fieldp *ModeratorCore, parallelism int) (m *ModeratorCore) {
	if fieldp != nil {
		m = fieldp
	} else {
		m = &ModeratorCore{}
	}
	if parallelism > 0 {
		m.parallelism = parallelism
	} else {
		m.parallelism = defaultParallelism
	}
	return
}

// TicketReturner is functional interface returned by [ModeratorCore.Ticket]
type TicketReturner interface {
	// returnTicket returns a ticket obtained from [ModeratorCore.Ticket]
	ReturnTicket()
}

// Ticket awaits and returns ticket
//   - always returns a ticket
//   - may block until ticket available
//   - the obtained ticket must be returned/released using either:
//   - — the returned object: m.Ticket().ReturnTicket()
//   - — [ModeratorCore.ReturnTicket]
//   - fair first-come-first-serve
//   - thread-safe
//
// Usage:
//
//	defer moderator.Ticket().ReturnTicket()
func (m *ModeratorCore) Ticket() (tickerReturner TicketReturner) {
	// Ticket always returns a ticket
	tickerReturner = m

	// initially try to get ticket at atomic performance
	//	- this fails when active == parallelism, ie. moderator lock-state
	if m.tryAtomicTicket() {
		return // got atomic ticket return
	}
	// thread encountered moderator in lock-state

	// seek ticket source: held ticket, atomic or ticketQueue
	if m.enterQueue() {
		return // got ticket in critical section return
	}

	// await ticket: blocks here for ticket from ticketQueue
	m.ticketQueue.Lock()

	return // ticket from ticketQueue return
}

// returnTicket returns a ticket obtained from [ModeratorCore.Ticket]
//   - thread-safe
func (m *ModeratorCore) ReturnTicket() {

	// current number of outstanding tickets
	var active = m.active.Load()
	// check for spurious ticket return
	if active == 0 {
		panic(perrors.NewPF("returning ticket without obtaining ticket"))
	}
	// attempt atomic ticket-return
	for active < m.parallelism {
		if m.active.CompareAndSwap(active, active-1) {
			return // ticket returned atomically return
		}
		active = m.active.Load()
	}
	// moderator in lock-state: active == parallelism
	//	- enter critical section
	defer m.lock.Lock().Unlock()

	// if there is a queue: give ticket to queue
	if m.waiting.Load() > 0 {
		m.ticketQueue.Unlock()
		m.waiting.Add(-1)
		return // ticket released to first thread in ticketQueue
	}

	for {

		// if a thread is progressing towards critical section,
		// hold ticket for it
		if m.seekers.Load() > m.heldTickets {
			m.heldTickets++
			return // ticket held for thread entering critical section return
		}

		// exit moderator lock-state
		var active = m.active.Load()
		if m.active.CompareAndSwap(active, active-1) {
			return // ticket returned atomically return
		}
	}
}

// Status: values may lack integrity
//   - parallelism: maximum number of tickets 1…
//   - active: current number of issued tickets 0–parallelism
//   - waiting: number of threads waiting for ticket
//   - thread-safe
func (m *ModeratorCore) Status() (parallelism, active, waiting int) {
	parallelism = m.parallelism
	active = m.active.Load()
	waiting = m.waiting.Load()
	return
}

// tryAtomicTicket tries to get ticket using atomics
//   - gotTicket true: a ticket was obtained using atomics
//   - gotTicket false: no ticket obtained, moderator is in lock mode
func (m *ModeratorCore) tryAtomicTicket() (gotTicket bool) {
	for {
		if tickets := m.active.Load(); tickets == m.parallelism {
			return // lock mode return: gotTicket false
		} else if gotTicket = m.active.CompareAndSwap(tickets, tickets+1); gotTicket {
			return // got atomic ticket return: gotTicket true
		}
	}
}

// enterQueue prepares thread to wait for ticket-queue lock
//   - gotTicket true: atomic ticket was obtained
//   - gotTicket false: enter the ticket queue
//   - —
//   - seekers must be non-zero
func (m *ModeratorCore) enterQueue() (gotTicket bool) {
	// signal that a thread is about to seek ticket in critical section
	m.seekers.Add(1)
	// enter critical section
	defer m.lock.Lock().Unlock()

	// thread no longer waiting for critical section
	m.seekers.Add(-1)

	// if there are held tickets, get that
	if gotTicket = m.heldTickets > 0; gotTicket {
		m.heldTickets--
		return // held ticket obtained return: gotTicket true
	}

	// if the moderator exited lock-state, get atomic ticket
	if gotTicket = m.tryAtomicTicket(); gotTicket {
		return // atomic ticket in critical section return: gotTicket true
	}

	// thread must enter ticketQueue
	var waiting = m.waiting.Add(1)

	// check for being first thread entering ticket queue
	if waiting > 1 || m.didFirst {
		return // other threads were already waiting: just join the queue
	}

	// first thread to wait
	m.didFirst = true
	// first thread to wait: lock the queue
	m.ticketQueue.Lock()

	return
}

// when tickets available: “available: 2(10)”
//   - parallelism is 10
//   - 10 - 2 = 8 threads operating
//   - when threads waiting “waiting 1(10)”
//   - 10 threads operating, 1 thread waiting
func (m *ModeratorCore) String() (s string) {
	var parallelism, active, waiting = m.Status()
	if active < parallelism {
		s = fmt.Sprintf("available: %d(%d)", parallelism-active, parallelism)
	} else {
		s = fmt.Sprintf("waiting: %d(%d)", waiting, parallelism)
	}
	return
}

const (
	// default is to allow 20 threads at a time
	defaultParallelism = 20
)
