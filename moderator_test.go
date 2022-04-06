/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	mttShortTime = 10 * time.Millisecond
)

func TestModerator(t *testing.T) {
	taskALaunched := "a"
	taskAExited := "A"
	goB := "d"
	doB := "b"
	doC := "c"
	goC := "¢"
	taskBExited := "B"
	taskCExited := "C"
	expectedSequence := "a¢dAcCbB"
	exitAIndex := strings.Index(expectedSequence, taskAExited)

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

	// instantiate
	count := 1
	ctx, cancel := context.WithCancel(context.Background())
	mo := NewModerator(1, ctx)

	// initial Status
	parallelism, active, waiting, shutdown := mo.Status()
	if parallelism != uint64(count) {
		t.Errorf("Parallelism %d expected %d", parallelism, count)
	}
	if active != 0 {
		t.Errorf("Parallelism %d expected %d", active, 0)
	}
	if waiting != 0 {
		t.Errorf("Waiting %d expected %d", waiting, 0)
	}
	if shutdown {
		t.Errorf("shutdown %t expected %t", shutdown, false)
	}

	// StateOne: status: one thread running, ni available tickets, no threads waiting
	extecpedA := 1
	extecpedW := 0
	wgADo := sync.WaitGroup{}
	wgADo.Add(1)
	wgAWait := sync.WaitGroup{}
	wgAWait.Add(1)
	wgAExit := sync.WaitGroup{}
	wgAExit.Add(1)
	// thread A goes into its Do function
	go func() {
		if err := mo.Do(func() (err error) {
			Append(taskALaunched)
			wgADo.Done()
			wgAWait.Wait()
			return
		}); err != nil {
			t.Logf("A err: %v", err)
			t.Fail()
		}
		Append(taskAExited)
		wgAExit.Done()
	}()
	wgADo.Wait()
	_, active, waiting, _ = mo.Status()
	if active != uint64(extecpedA) {
		t.Errorf("State One: active expected: %d actual: %d", active, extecpedA)
	}
	if waiting != uint64(extecpedW) {
		t.Errorf("State One: waiting expected: %d actual: %d", waiting, extecpedW)
	}

	// state two: 2 threads in queue
	extecpedW = 2
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
	// thread B waiting
	go BC(goB, doB, taskBExited)
	// thread C waiting
	go BC(goC, doC, taskCExited)
	// wait for threads to enter cond.Wait()
	time.Sleep(mttShortTime)
	_, active, waiting, _ = mo.Status()
	if active != uint64(extecpedA) {
		t.Errorf("State One: active expected: %d actual: %d", active, extecpedA)
	}
	if waiting != uint64(extecpedW) {
		t.Errorf("State One: waiting expected: %d actual: %d", waiting, extecpedW)
	}

	// state three: 1 thread in queue
	extecpedW = 1
	wgAWait.Done()
	wgAExit.Wait()
	// wait for a thread to exit cond.Wait()
	time.Sleep(mttShortTime)
	_, active, waiting, _ = mo.Status()
	if active != uint64(extecpedA) {
		t.Errorf("State One: active expected: %d actual: %d", active, extecpedA)
	}
	if waiting != uint64(extecpedW) {
		t.Errorf("State One: waiting expected: %d actual: %d", waiting, extecpedW)
	}

	// state 4: 1 ticket available
	extecpedA = 0
	extecpedW = 0
	wgBCWait.Done()
	wgBCExit.Wait()
	_, active, waiting, _ = mo.Status()
	if active != uint64(extecpedA) {
		t.Errorf("State One: active expected: %d actual: %d", active, extecpedA)
	}
	if waiting != uint64(extecpedW) {
		t.Errorf("State One: waiting expected: %d actual: %d", waiting, extecpedW)
	}

	seq := Append("") // get result
	// expected sequence: a¢dAcCbB
	// ¢d and cCbB may flip
	if len(seq) != len(expectedSequence) ||
		seq[0:1] != expectedSequence[0:1] ||
		seq[exitAIndex:exitAIndex+1] != expectedSequence[exitAIndex:exitAIndex+1] {
		t.Logf("Bad sequence: expected: %s actual: %s", expectedSequence, seq)
		t.Fail()
	}

	cancel()
	_, _, _, shutdown = mo.Status()
	if !shutdown {
		t.Errorf("shutdown %t expected %t", shutdown, true)
	}

}
