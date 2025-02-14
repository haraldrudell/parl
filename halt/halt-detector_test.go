/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package halt

import (
	"context"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/g0"
	"github.com/haraldrudell/parl/g0/g0debug"
)

func TestHaltDetector(t *testing.T) {
	const (
		// timeout to wait for a halt report to be produced
		timeout1s = time.Second
		// timeout to wait for goError
		goErrorTimeout = time.Millisecond
		// 0 ns means a report is issued with certainty after the first interval
		reportingThreshold time.Duration = 0
		// 0 ns means a minimum interval is used
		intervalToUse time.Duration = 0
	)
	var (
		// the start time for test to compare with halt report time
		start       = time.Now()
		maxDuration time.Duration
	)

	var (
		ch                parl.IterableSource[HaltReport]
		goGroup           parl.GoGroup
		haltReport        HaltReport
		end               time.Time
		goError           parl.GoError
		hasValue, hadExit bool
		goErrorSource     parl.ClosableSource1[parl.GoError]
		endCh             parl.AwaitableCh
	)

	// Ch() Thread()
	var haltDetector *HaltDetector = NewHaltDetector2(
		NoHaltFieldp,
		reportingThreshold,
		intervalToUse,
		MonotonicYes,
	)

	// Ch should return channel for HaltReports
	ch = haltDetector.Ch()
	if ch == nil {
		t.Error("Ch nil")
	}

	// haltDetector.Thread should produce a halt report within 1 second
	//	- start halt detector with a thread-group that can cancel it
	goGroup = g0.NewGoGroup(context.Background())
	go haltDetector.Thread(goGroup.Go())
	var timer = time.NewTimer(timeout1s)
	defer timer.Stop()
	// await report
	select {
	case <-ch.DataWaitCh():
		haltReport, hasValue = ch.Get()
		if !hasValue {
			t.Fatal("has data but no halt report")
		}
	case <-timer.C:
		t.Fatalf("no halt report within %s", timeout1s)
	}
	end = time.Now()
	// a halt report was received within 1 second

	// halt report should be consistent
	if haltReport.Number != 1 {
		t.Errorf("haltReport.N not 1: %d", haltReport.Number)
	}
	if haltReport.Timestamp.Before(start) {
		t.Error("haltReport.T Before start")
	}
	if haltReport.Timestamp.After(end) {
		t.Error("haltReport.T After end")
	}
	maxDuration = end.Sub(start)
	if haltReport.Duration < 0 || haltReport.Duration > maxDuration {
		t.Errorf("haltReport.D bad %s", haltReport.Duration)
	}

	// context cancel should shutdown thread
	// shutdown
	//	- exit Thread by canceling context
	goGroup.Cancel()
	// want to be able to wait with timeout
	//	- therefore use the source directly
	goErrorSource = goGroup.GoError()
	endCh = goErrorSource.EmptyCh(parl.CloseAwaiter)
	timer = time.NewTimer(goErrorTimeout)
	defer timer.Stop()
	for {
		select {
		case <-endCh: // error channel was read to end
		case <-goErrorSource.DataWaitCh():
			goError, hasValue = goErrorSource.Get()
			if !hasValue {
				t.Fatal("has value but hasValue false")
			}
			if goError != nil &&
				!hadExit &&
				goError.ErrContext() == parl.GeExit &&
				goError.Err() == nil {
				hadExit = true
				continue
			}
			t.Fatalf("unexpected goError: %s", g0debug.GoErrorDump(goError))
		case <-timer.C:
			t.Fatalf("goError read timed out: %s", goErrorTimeout)
		}
		break
	}
	timer.Stop()
	// goGroup error channel closed
}
