/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"sync"
	"testing"
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

func TestNewOnTicker(t *testing.T) {
	//t.Error("Logging on")
	const (
		// period is a short period but
		// long enough to detect hang
		period    = time.Millisecond
		shortTime = time.Millisecond
	)

	var (
		// timestamp before starting the timer
		t0 time.Time
		// timestamp afteer first trig
		t1 time.Time
		// first timestamp received from timer
		actualTime time.Time
		// second timestamp received from timer
		actual2 time.Time
		// time between subsequent ticks
		duration     time.Duration
		isThreadExit chan struct{}
		timer        *time.Timer
	)

	// Stop()
	var onTicker *OnTicker

	// initial period-aligning tick should be beetween t0 and t1
	//	- thread will be created and exit
	t0 = time.Now()
	onTicker = NewOnTicker(period, nil)
	actualTime = <-onTicker.C
	t1 = time.Now()

	// t0: 2025-05-15 07:53:31.228868000-07:00
	// actualTime: 2025-05-15 07:53:31.229040834-07:00
	// t1: 2025-05-15 07:53:31.229064000-07:00
	t.Logf("t0: %s\nactualTime: %s\nt1: %s",
		t0.Format(cyclebreaker.Rfc3339ns),
		actualTime.Format(cyclebreaker.Rfc3339ns),
		t1.Format(cyclebreaker.Rfc3339ns),
	)

	if !actualTime.After(t0) || !actualTime.Before(t1) {
		t.Errorf("FAIL bad time received on channel\nt0: %s\nt: %s\nt1: %s",
			t0.Format(cyclebreaker.Rfc3339ns),
			actualTime.Format(cyclebreaker.Rfc3339ns),
			t1.Format(cyclebreaker.Rfc3339ns),
		)
	}

	// second tick: thread should be exited
	isThreadExit = make(chan struct{})
	go waitThread(&onTicker.wg, isThreadExit)
	actual2 = <-onTicker.C
	timer = time.NewTimer(shortTime)
	defer timer.Stop()
	select {
	case <-timer.C:
		t.Errorf("FAIL thread exit timeout")
	case <-isThreadExit:
	}

	// actual2 should be one period after actualTimer
	duration = actual2.Sub(actualTime)
	if duration < period/2 {
		t.Errorf("duration too short: %s exp %s", duration, period)
	} else if duration >= 2*period {
		t.Errorf("duration too long: %s exp %s", duration, period)
	}
}

// waitThread awaits waitgroup trig
func waitThread(threadExit *sync.WaitGroup, isThreadExit chan struct{}) {
	defer close(isThreadExit)

	threadExit.Wait()
}
