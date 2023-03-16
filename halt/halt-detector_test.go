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
)

func TestHaltDetector(t *testing.T) {
	var start = time.Now()
	var minD = 500 * time.Microsecond

	var haltDetector *HaltDetector
	var ch <-chan *HaltReport
	var goGroup parl.GoGroup
	var haltReport *HaltReport
	var end time.Time
	var goError parl.GoError
	var ok bool

	// create
	haltDetector = NewHaltDetector()
	ch = haltDetector.Ch()

	// run thread
	goGroup = g0.NewGoGroup(context.Background())
	go haltDetector.Thread(goGroup.Go())

	// check report
	haltReport = <-ch
	end = time.Now()
	if haltReport.N != 1 {
		t.Errorf("haltReport.N not 1: %d", haltReport.N)
	}
	if haltReport.T.Before(start) {
		t.Error("haltReport.T Before start")
	}
	if haltReport.T.After(end) {
		t.Error("haltReport.T After end")
	}
	if haltReport.D < minD {
		t.Errorf("haltReport.D less than %s", minD)
	}
	// shutdown
	goGroup.Cancel()
	// receive thread exit
	if goError, ok = <-goGroup.Ch(); !ok {
		t.Error("goGroup.Ch closed")
	}
	if goError.Err() != nil || goError.ErrContext() != parl.GeExit {
		t.Errorf("goGroup bad goError: %s", goError.String())
	}
	if goError, ok = <-goGroup.Ch(); ok {
		t.Errorf("goGroup.Ch did not close: %s", goError.String())
	}
}
