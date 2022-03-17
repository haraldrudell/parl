/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestModerator(t *testing.T) {
	doA := "a"
	exitA := "A"
	goB := "d"
	doB := "b"
	doC := "c"
	goC := "¢"
	exitB := "B"
	exitC := "C"
	expectedSequence := "a¢dAcCbB"
	exitAIndex := strings.Index(expectedSequence, exitA)

	Append := func() func(string) string {
		sequence := ""
		seqLock := sync.Mutex{}
		return func(sIn string) (sOut string) {
			seqLock.Lock()
			defer seqLock.Unlock()
			sequence += sIn
			return sequence
		}
	}()

	// new state
	expectedStatusNew := "available: 1(1)"
	mo := NewModerator(1)
	actual := mo.String()
	if actual != expectedStatusNew {
		t.Logf("New status: expected: %s actual: %s", expectedStatusNew, actual)
		t.Fail()
	}

	// state one: one thread running, no tickets left
	expectedStatusOne := "waiting: 0(1)"
	wgADo := sync.WaitGroup{}
	wgADo.Add(1)
	wgAWait := sync.WaitGroup{}
	wgAWait.Add(1)
	wgAExit := sync.WaitGroup{}
	wgAExit.Add(1)
	go func() {
		if err := mo.Do(func() (err error) {
			Append(doA)
			wgADo.Done()
			wgAWait.Wait()
			return
		}); err != nil {
			t.Logf("A err: %v", err)
			t.Fail()
		}
		Append(exitA)
		wgAExit.Done()
	}()
	wgADo.Wait()
	actual = mo.String()
	if actual != expectedStatusOne {
		t.Logf("State One: expected: %s actual: %s", expectedStatusOne, actual)
		t.Fail()
	}

	// state two: 2 threads in queue
	expectedStatusTwo := "waiting: 2(1)"
	wgBCWait := sync.WaitGroup{}
	wgBCWait.Add(1)
	wgBCExit := sync.WaitGroup{}
	wgBCExit.Add(2)
	BC := func(goo, do, exit string) {
		Append(goo)
		if err := mo.Do(func() (err error) {
			Append(do)
			wgBCWait.Wait()
			return
		}); err != nil {
			t.Logf("%s err: %v", do, err)
		}
		Append(exit)
		wgBCExit.Done()
	}
	go BC(goB, doB, exitB)
	go BC(goC, doC, exitC)
	for {
		actual = mo.String() // will hang if something wrong
		if expectedStatusTwo == actual {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// state three: 1 thread in queue
	expectedStatusThree := "waiting: 1(1)"
	wgAWait.Done()
	wgAExit.Wait()
	actual = mo.String()
	if actual != expectedStatusThree {
		t.Logf("Status Three: expected: %s actual: %s", expectedStatusThree, actual)
		t.Fail()
	}

	// state 4: 1 ticket available
	expectedStatusFour := "available: 1(1)"
	wgBCWait.Done()
	wgBCExit.Wait()
	actual = mo.String()
	if actual != expectedStatusFour {
		t.Logf("Status Four: expected: %s actual: %s", expectedStatusFour, actual)
		t.Fail()
	}

	seq := Append("")
	// expected sequence: a¢dAcCbB
	// ¢d and cCbB may flip
	if len(seq) != len(expectedSequence) ||
		seq[0:1] != expectedSequence[0:1] ||
		seq[exitAIndex:exitAIndex+1] != expectedSequence[exitAIndex:exitAIndex+1] {
		t.Logf("Bad sequence: expected: %s actual: %s", expectedSequence, seq)
		t.Fail()
	}
}
