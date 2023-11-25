/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

const (
	// default is to allow 20 threads at a time
	defaultParallelism = 20
)

// ModeratorCore invokes functions at a limited level of parallelism
//   - ModeratorCore is a ticketing system
//   - ModeratorCore does not have a cancel feature
//   - during low contention atomic performance
//   - during high-contention lock performance
//
// Usage:
//
//	m := NewModeratorCore(20, ctx)
//	defer m.Ticket()() // waiting here for a ticket
//	// got a ticket!
//	…
//	return or panic // ticket automatically returned
//	m.String() → waiting: 2(20)
type ModeratorCore struct {
	// parallelism is the maximum number of outstanding tickets
	parallelism uint64
	// number of issued tickets
	//	- if less than parallelism:
	//	- — moderator is in atomic mode, ie.
	//	- —	tickets obtained by atomic access only
	//	- if equal to parallelism:
	//	- — moderator is in lock mode. ie.
	//	- — tickets are transfered orderly using queue,
	//		waiting and transferBehindLock
	active atomic.Uint64
	// lock used when moderator in lock mode
	//	- treads use the cond with waiting and transferBehindLock
	//	- orderly first-come-first-served
	queue sync.Cond
	// number of threads waiting for a ticket
	//	- behind lock
	//	- atomic so Status can read
	waiting atomic.Uint64
	// transferBehindLock facilitates locked ticket transfer
	//	- behind lock
	//	- atomic so it can be inspected
	transferBehindLock atomic.Uint64
}

// moderatorCore is a parl-private version of ModeratorCore
type moderatorCore struct {
	*ModeratorCore
}

// NewModerator creates a new Moderator used to limit parallelism
func NewModeratorCore(parallelism uint64) (m *ModeratorCore) {
	if parallelism < 1 {
		parallelism = defaultParallelism
	}
	return &ModeratorCore{
		parallelism: parallelism,
		queue:       *sync.NewCond(&sync.Mutex{}),
	}
}

// Ticket returns a ticket possibly blocking until one is available
//   - Ticket returns the function for returning the ticket
//
// Usage:
//
//	defer moderator.Ticket()()
func (m *ModeratorCore) Ticket() (returnTicket func()) {
	returnTicket = m.returnTicket

	// try available ticket at atomic performance
	for {
		if tickets := m.active.Load(); tickets == m.parallelism {
			break // it’s lock mode
		} else if m.active.CompareAndSwap(tickets, tickets+1) {
			return // got atomic ticket return
		}
	}

	// enter lock mode
	m.queue.L.Lock()
	defer m.queue.L.Unlock()
	defer m.lastWaitCheck()

	// critial section: ticket loop
	var isWaiting bool
	for {

		// attempt atomic ticket
		for {
			if tickets := m.active.Load(); tickets == m.parallelism {
				break // still lock mode
			} else if m.active.CompareAndSwap(tickets, tickets+1) {
				return // got atomic ticket return
			}
		}

		// attempt transfer-behind-lock ticket
		if m.transferBehindLock.Load() > 0 {
			m.transferBehindLock.Add(math.MaxUint64)
			return // ticket transfer successful return
		}

		// wait for ticket to become available
		if !isWaiting {
			isWaiting = true
			m.waiting.Add(1)
			defer m.waiting.Add(math.MaxUint64)
		}
		// blocks here
		m.queue.Wait()
	}
}

// lastWaitCheck prevents tickets from getting stuck as transfers
//   - invoked while holding lock
//   - this can happen if 1 thread is waiting and multiple threads transfer tickets
func (m *ModeratorCore) lastWaitCheck() {
	if m.waiting.Load() > 0 {
		return // more threads are waiting
	}
	var transfers = m.transferBehindLock.Load()
	if transfers == 0 {
		return // no extra transfers available return
	}

	// put extra transfers in atomic tickets
	m.transferBehindLock.Store(0)
	m.active.Add(math.MaxUint64 - transfers + 1)
}

// returnTicket returns a ticket obtained by Ticket
func (m *ModeratorCore) returnTicket() {

	// attempt ticket-return atomically
	for {
		if tickets := m.active.Load(); tickets == m.parallelism {
			break // lock mode: use transfer-ticket
		} else if m.active.CompareAndSwap(tickets, tickets-1) {
			return // ticket returned atomically return
		}
	}

	// return ticket using transfer behind lock
	m.queue.L.Lock()
	defer m.queue.L.Unlock()

	// if no thread waiting, return atomically
	if m.waiting.Load() == 0 {
		m.active.Add(math.MaxUint64)
		return // atomic transfer complete return
	}

	// if thread waiting, do ticket transfer
	m.transferBehindLock.Add(1)
	m.queue.Signal() // signal while holding lock
}

// Status: values may lack integrity
func (m *ModeratorCore) Status() (parallelism, active, waiting uint64) {
	parallelism = m.parallelism
	active = m.active.Load()
	waiting = m.waiting.Load()
	return
}

// when tickets available: “available: 2(10)”
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
