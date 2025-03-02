/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestModeratorCore(t *testing.T) {
	//t.Error("Logging On")
	const (
		parallelism        = 1
		waitForTicketQueue = time.Millisecond
	)
	var (
		exp1_1Ticket = moderatorState{
			parallelism: parallelism,
			active:      1,
		}
		exp2_Ticket2Wait = moderatorState{
			parallelism: parallelism,
			active:      1,
			waiting:     1,
			didFirst:    true,
		}
		exp3_1Returned = moderatorState{
			parallelism: parallelism,
			active:      1,
			didFirst:    true,
		}
		exp4_2returned = moderatorState{
			parallelism: parallelism,
			didFirst:    true,
		}
	)

	var (
		ticketReturner TicketReturner
		actual         moderatorState
	)

	// Ticket() ReturnTicket() Status() String()
	var moderator *ModeratorCore = NewModeratorCore(parallelism)

	// ticketReturner should be non-nil
	ticketReturner = moderator.Ticket()
	if ticketReturner == nil {
		t.Fatal("FAIL ticketReturner nil")
	}
	// one ticket should be atomic access
	actual = moderatorStatus(moderator)
	if actual != exp1_1Ticket {
		t.Errorf("FAIL one ticket:\n%#v exp:\n%#v", actual, exp1_1Ticket)
	}

	// ticket two should wait in ticketQueue
	var ticketGetter = newTicketGetter(moderator, t)
	ticketGetter.createThread()
	// wait for created ticketThread running
	<-ticketGetter.threadReady.Ch()
	// wait 1 ms for thread to reach m.queue.Wait
	//	- read atomics
	//	- enter-exit critical section
	//	- hold at ticketQueue mutex
	<-time.NewTimer(waitForTicketQueue).C
	// ticket two should not have gotten ticket
	if ticketGetter.haveTicket.Load() {
		t.Errorf("FAIL haveTicket true")
	}
	// ticket two should be in ticketQueue
	actual = moderatorStatus(moderator)
	if actual != exp2_Ticket2Wait {
		t.Errorf("FAIL ticket2 waiting:\n%#v exp:\n%#v", actual, exp2_Ticket2Wait)
	}

	// return first ticket should release ticket two
	ticketReturner.ReturnTicket()
	// ticket two should no longer be blocked in ticketQueue
	ticketGetter.threadIsExit.Wait()
	// ticket two should have ticket
	if !ticketGetter.haveTicket.Load() {
		t.Errorf("FAIL haveTicket false")
	}
	// moderator state should match
	actual = moderatorStatus(moderator)
	if actual != exp3_1Returned {
		t.Errorf("FAIL after return1:\n%#v exp:\n%#v", actual, exp3_1Returned)
	}
	// ticket two shold have ticketReturner
	if ticketGetter.ticketReturner == nil {
		t.Fatal("FAIL returnTicket2 nil")
	}

	// return second ticket should match moderator state
	ticketGetter.ticketReturner.ReturnTicket()
	actual = moderatorStatus(moderator)
	if actual != exp4_2returned {
		t.Errorf("FAIL after return2:\n%#v exp:\n%#v", actual, exp4_2returned)
	}

	// end of test
}

// moderatorStatus holds moderator state during testing
type moderatorState struct {
	parallelism, active, seekers, waiting, heldTickets int
	didFirst                                           bool
}

// moderatorStatus returns snapshot of moderator state
func moderatorStatus(m *ModeratorCore) (s moderatorState) {
	s = moderatorState{}
	s.parallelism, s.active, s.waiting = m.Status()
	s.seekers = m.seekers.Load()
	defer m.lock.Lock().Unlock()

	s.heldTickets, s.didFirst = m.heldTickets, m.didFirst
	return
}

// ticketGetter provides a thread obtaining a ticket from a moderator
type ticketGetter struct {
	// moderatorCore is the moderator being tested
	moderatorCore *ModeratorCore
	t             *testing.T
	// thread status
	threadReady, threadIsExit WaitGroupCh
	// true once thread has obtained a ticket
	haveTicket atomic.Bool
	// valid once haveTicket true
	ticketReturner TicketReturner
}

// newTicketGetter returns an object obtaining tickets in a thread
func newTicketGetter(m0 *ModeratorCore, t *testing.T) (m *ticketGetter) {
	return &ticketGetter{
		moderatorCore: m0,
		t:             t,
	}
}

// createThread creates a thread that obtains a ticket, then exits
func (m *ticketGetter) createThread() {
	m.threadReady.Add(1)
	m.threadIsExit.Add(1)
	go m.ticketThread()
}

// ticketThread obtains a ticket and exits
func (m *ticketGetter) ticketThread() {
	var t = m.t
	defer m.threadIsExit.Done()

	// signal thread created and running
	m.threadReady.Done()

	t.Log("thread invoking Ticket")
	m.ticketReturner = m.moderatorCore.Ticket()

	t.Log("thread got ticket")
	m.haveTicket.Store(true)
}
