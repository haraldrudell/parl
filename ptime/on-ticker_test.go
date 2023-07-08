/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package ptime

import (
	"testing"
	"time"

	"github.com/haraldrudell/parl/internal/cyclebreaker"
)

func TestNewOnTicker(t *testing.T) {
	var period = time.Millisecond

	// start a ticker that has a thread waiting for period alignment
	var t0 = time.Now()
	var ticker = NewOnTicker(period, nil)

	// wait for first tick: thread will exit
	var tx = <-ticker.C
	var t1 = time.Now()
	if !tx.After(t0) || !tx.Before(t1) {
		t.Errorf("bad time received on channel\nt0: %s\nt: %s\nt1: %s",
			t0.Format(cyclebreaker.Rfc3339ns),
			tx.Format(cyclebreaker.Rfc3339ns),
			t1.Format(cyclebreaker.Rfc3339ns),
		)
	}

	// second tick: thread should be exited
	var ty = <-ticker.C
	if !ticker.onTickp.isThreadExit.Load() {
		t.Error("thread did not exit")
	}

	var duration = ty.Sub(tx)
	if duration < period/2 {
		t.Errorf("duration too short: %s exp %s", duration, period)
	} else if duration >= 2*period {
		t.Errorf("duration too long: %s exp %s", duration, period)
	}
}
