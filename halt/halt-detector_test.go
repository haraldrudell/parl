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
	var (
		start = time.Now()
		minD  = 500 * time.Microsecond
	)

	var (
		haltDetector *HaltDetector
		ch           parl.Source1[*HaltReport]
		goGroup      parl.GoGroup
		haltReport   *HaltReport
		end          time.Time
		goError      parl.GoError
		ok, hasValue bool
	)

	// create
	haltDetector = NewHaltDetector()
	ch = haltDetector.Ch()

	// run thread
	goGroup = g0.NewGoGroup(context.Background())
	go haltDetector.Thread(goGroup.Go())

	// check report
	<-ch.DataWaitCh()
	haltReport, hasValue = ch.Get()
	_ = hasValue
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
	if goError, ok = parl.AwaitValue(goGroup.GoError()); !ok {
		t.Error("goGroup.Ch closed")
	}
	if goError.Err() != nil || goError.ErrContext() != parl.GeExit {
		t.Errorf("goGroup bad goError: %s", goError.String())
	}
	if goError, ok = parl.AwaitValue(goGroup.GoError()); ok {
		t.Errorf("goGroup.Ch did not close: %s", goError.String())
	}
}
