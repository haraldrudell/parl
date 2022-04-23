/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
	"time"
)

const ocShort = 10 * time.Millisecond

func TestOnceChan(t *testing.T) {
	var wg WaitGroup
	var onceChan OnceChan

	// have one thread waiting
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-onceChan.Done()
	}()
	time.Sleep(ocShort)
	expWaiting := 1
	adds, dones := wg.Counters()
	actual := adds - dones
	if actual != expWaiting {
		t.Errorf("Bad waiting: %d expected %d", actual, expWaiting)
	}
	if onceChan.IsDone() {
		t.Error("onceChan should not be isDone")
	}
	if onceChan.Err() != nil {
		t.Error("onceChan should not have Err()")
	}

	// Done state
	onceChan.Cancel()
	if !onceChan.IsDone() {
		t.Error("onceChan should be isDone")
	}
	if onceChan.Err() == nil {
		t.Error("onceChan should have Err()")
	}
	wg.Wait()
	_, ok := <-onceChan.Done()
	if ok {
		t.Error("onceChan channel not closed")
	}
}
