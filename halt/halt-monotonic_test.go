/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package halt

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/g0"
)

// ITEST= go test -v -count 1 -run ^TestMonotonic$ github.com/haraldrudell/parl/halt
func TestMonotonic(t *testing.T) {

	// this integration test only runs if environment variable ITEST= is set
	if _, ok := os.LookupEnv("ITEST"); !ok {
		t.Skip("skip because ITEST not set")
	}

	const (
		// cancel halt detector after 5 seconds
		runFor = 5 * time.Second
		// 0 ns means a report is issued with certainty after the first interval
		reportingThreshold time.Duration = 0
		// 0 ns means a minimum interval is used
		intervalToUse time.Duration = 0
	)

	var (
		goGroup                      parl.GoGroup
		timer                        *time.Timer
		threadCancel, threadComplete = make(chan struct{}), make(chan struct{})
		hTester                      *haltTester
	)

	var haltDetector *HaltDetector = NewHaltDetector2(
		NoHaltFieldp,
		reportingThreshold,
		intervalToUse,
		MonotonicYes,
	)

	// create thread
	goGroup = g0.NewGoGroup(context.Background())
	t.Logf("creating thread to run for %s", runFor)
	t.Logf("isMonotonic %t threshold %s interval %s",
		haltDetector.isMonotonic,
		haltDetector.reportingThreshold,
		haltDetector.interval,
	)
	go haltDetector.Thread(goGroup.Go())

	// launch cancel thread
	timer = time.NewTimer(runFor)
	defer timer.Stop()
	hTester = newHaltTester(
		timer.C,
		goGroup, threadCancel, threadComplete,
		t,
	)
	defer hTester.await()
	go hTester.cancelContext()
	defer close(threadCancel)

	// print reports
	for haltReport := range haltDetector.Ch().Seq {
		t.Log(haltReport.String())
	}
	t.Log("report channel closed")
}

type haltTester struct {
	C                            <-chan time.Time
	g                            parl.GoGroup
	threadCancel, threadComplete chan struct{}
	t                            *testing.T
}

func newHaltTester(
	C <-chan time.Time,
	g parl.GoGroup,
	threadCancel, threadComplete chan struct{},
	t *testing.T,
) (h *haltTester) {
	return &haltTester{
		C:              C,
		g:              g,
		threadCancel:   threadCancel,
		threadComplete: threadComplete,
		t:              t,
	}
}

func (h *haltTester) cancelContext() {
	var t = h.t
	defer close(h.threadComplete)
	select {
	case <-h.threadCancel:
		return
	case <-h.C:
	}
	t.Log("context cancel")
	h.g.Cancel()
}

func (h *haltTester) await() { <-h.threadComplete }
