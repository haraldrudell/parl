/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	shortTime = time.Millisecond
)

type ModeratorStatus struct {
	parallelism, active, waiting, transferBehindLock uint64
}

func NewModeratorStatus(m *ModeratorCore) (status *ModeratorStatus) {
	s := ModeratorStatus{}
	s.parallelism, s.active, s.waiting = m.Status()
	s.transferBehindLock = m.transferBehindLock.Load()
	return &s
}

type ModeratorLocker struct {
	isReady, isDone sync.WaitGroup
	mo              *ModeratorCore
	haveTicket      atomic.Bool
	returnTicket    func()
	t               *testing.T
}

func NewModeratorLocker(mo *ModeratorCore, t *testing.T) (m *ModeratorLocker) {
	return &ModeratorLocker{mo: mo, t: t}
}
func (m *ModeratorLocker) Ticket() {
	m.isReady.Add(1)
	m.isDone.Add(1)
	go m.ticket()
}
func (m *ModeratorLocker) ticket() {
	defer m.isDone.Done()

	var t = m.t
	m.isReady.Done()
	t.Log("thread invoking Ticket")
	m.returnTicket = m.mo.Ticket()
	t.Log("thread got ticket")
	m.haveTicket.Store(true)
}

func TestModeratorCore(t *testing.T) {
	var parallelism = uint64(1)
	var exp1_1Ticket = ModeratorStatus{
		parallelism: parallelism,
		active:      1,
	}
	var exp2_Ticket2Wait = ModeratorStatus{
		parallelism: parallelism,
		active:      1,
		waiting:     1,
	}
	var exp3_1Returned = ModeratorStatus{
		parallelism: parallelism,
		active:      1,
	}
	var exp4_2returned = ModeratorStatus{
		parallelism: parallelism,
	}

	var moderator *ModeratorCore
	var returnTicket1 func()
	var actual ModeratorStatus

	// one thread should be atomic access
	moderator = NewModeratorCore(parallelism)
	returnTicket1 = moderator.Ticket()
	if returnTicket1 == nil {
		t.Fatal("returnTicket1 nil")
	}
	actual = *NewModeratorStatus(moderator)
	if actual != exp1_1Ticket {
		t.Errorf("1: one ticket:\n%#v exp:\n%#v", actual, exp1_1Ticket)
	}

	// thread2 should use lock mode and block
	var ml = NewModeratorLocker(moderator, t)
	ml.Ticket()
	ml.isReady.Wait() // wait for thread Ticket invocation
	// wait 1 ms for thread to reach m.queue.Wait
	<-time.NewTimer(shortTime).C
	if ml.haveTicket.Load() {
		t.Errorf("ml.haveTicket true")
	}
	actual = *NewModeratorStatus(moderator)
	if actual != exp2_Ticket2Wait {
		t.Errorf("2: ticket2 waiting:\n%#v exp:\n%#v", actual, exp2_Ticket2Wait)
	}

	// return first ticket
	returnTicket1()
	ml.isDone.Wait()
	if !ml.haveTicket.Load() {
		t.Errorf("ml.haveTicket false")
	}
	actual = *NewModeratorStatus(moderator)
	if actual != exp3_1Returned {
		t.Errorf("3: after return1:\n%#v exp:\n%#v", actual, exp3_1Returned)
	}
	if ml.returnTicket == nil {
		t.Fatal("returnTicket2 nil")
	}

	// return second ticket
	ml.returnTicket()
	actual = *NewModeratorStatus(moderator)
	if actual != exp4_2returned {
		t.Errorf("4: after return2:\n%#v exp:\n%#v", actual, exp4_2returned)
	}
}
