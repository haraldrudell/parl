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

// ITEST= go test -v -count 1 -run ^TestThreshold$ github.com/haraldrudell/parl/halt
func TestThreshold(t *testing.T) {

	// this integration test only runs if environment variable ITEST= is set
	if _, ok := os.LookupEnv("ITEST"); !ok {
		t.Skip("skip because ITEST not set")
	}

	const (
		// 1 ms halt typically occurrs within a few seconds
		reportingThreshold = time.Millisecond
		// something less than threshold
		intervalToUse = reportingThreshold / 2
	)

	var (
		goGroup parl.GoGroup
		err     error
		g       parl.GoResult
	)

	var haltDetector *HaltDetector = NewHaltDetector2(
		NoHaltFieldp,
		reportingThreshold,
		intervalToUse,
	)

	// create thread-group
	goGroup = g0.NewGoGroup(context.Background())
	defer goGroup.Wait()
	// must read error channel toend, so use separate thread
	g = parl.NewGoResult()
	defer errIsFatal(&err, t)
	defer g.ReceiveError(&err)
	go g0.Reader(parl.NoSTReader, parl.NoErrorSink1, t.Logf, goGroup, g)
	defer goGroup.Cancel()

	// create thread
	t.Logf("creating thread to detect halt of at least %s", reportingThreshold)
	t.Log("typically happens in a matter of seconds")
	t.Log("^C to cancel")
	t.Logf("isMonotonic %t interval %s",
		haltDetector.isMonotonic,
		haltDetector.interval,
	)
	go haltDetector.Thread(goGroup.Go())

	// print first report
	for haltReport := range haltDetector.Ch().Seq {
		t.Log(haltReport.String())
		t.Log("context cancel")
		break
	}
}

// errIsFatal examines *errp and invokes t.Errorf on error
//   - “reader error …”
func errIsFatal(errp *error, t *testing.T) {
	var err = *errp
	if err == nil {
		return // ok return
	}
	t.Errorf("reader error %s", err)
}
