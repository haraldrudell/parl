/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/g0"
)

func Test_newRateCounter(t *testing.T) {
	//t.Error("Logging on")
	var (
		messagePeriod = "period must be positive"
		periodZero    time.Duration
		waitOneSecond = time.Second
	)

	var (
		err             error
		counterCounters *Counters
		threadGroup     parl.GoGroup
		goError         parl.GoError
		timer           *time.Timer
		isTimeout       bool
	)

	// GetOrCreateCounter GetOrCreateDatapoint
	var counters parl.Counters
	var reset = func() {
		// g0 nil: no counter threads allowed
		threadGroup = g0.NewGoGroup(context.Background())
		counters = newCounters(threadGroup.Go())
		counterCounters, _ = counters.(*Counters)
	}

	// newRateCounter zero period panic
	reset()
	// AddTask creates a Go ptime.OnTickerThread that will panic
	// due to period zero
	counterCounters.AddTask(periodZero, newRateCounter())
	// threadGroup should have a fatal thread exit
	//	- once the goroutine starts and panics
	//	- counters holds one parl.Go as GoGen
	//	- there is no access to the Go created for OnTickerThread
	//	- the thread-group will send errors but not close its error stream
	//	- any awaitable may or may not trigger
	//	- meaning must wait with a timeout for a goroutine to start and panic

	// TODO 240526 refactor counter package to support shutdown
	//	- invoke Done on any Go
	//	- shutdown timer threads

	timer = time.NewTimer(waitOneSecond)
	defer timer.Stop()
	isTimeout = false
	for {
		select {
		case <-threadGroup.GoError().DataWaitCh():
			var hasValue bool
			if goError, hasValue = threadGroup.GoError().Get(); !hasValue {
				continue
			}
		case <-timer.C:
			isTimeout = true
		}
		break
	}
	if isTimeout {
		t.Fatal("gogroup goerror timeout")
	}
	err = goError.Err()
	if err == nil {
		t.Error("RecoverInvocationPanic exp panic missing")
	} else if !strings.Contains(err.Error(), messagePeriod) {
		t.Errorf("RecoverInvocationPanic bad err: %q exp %q", err.Error(), messagePeriod)
	}
}
