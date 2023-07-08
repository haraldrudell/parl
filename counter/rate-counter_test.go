/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"context"
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/g0"
)

func Test_newRateCounter(t *testing.T) {
	messagePeriod := "period must be positive"

	var err error

	var counters parl.Counters
	var counterCounters *Counters

	// g0 nil: no counter threads allowed
	var threadGroup = g0.NewGoGroup(context.Background())
	counters = newCounters(threadGroup.Go())
	counterCounters, _ = counters.(*Counters)

	// newRateCounter zero period panic
	counterCounters.AddTask(0, newRateCounter())
	goError := <-threadGroup.Ch()
	err = goError.Err()
	if err == nil {
		t.Error("RecoverInvocationPanic exp panic missing")
	} else if !strings.Contains(err.Error(), messagePeriod) {
		t.Errorf("RecoverInvocationPanic bad err: %q exp %q", err.Error(), messagePeriod)
	}
}
