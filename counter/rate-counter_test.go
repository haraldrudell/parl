/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"strings"
	"testing"

	"github.com/haraldrudell/parl"
)

func Test_newRateCounter(t *testing.T) {
	messagePeriod := "period must be positive"

	var err error

	var counters parl.Counters
	var counterCounters *Counters

	// g0 nil: no counter threads allowed
	counters = newCounters(nil)
	counterCounters, _ = counters.(*Counters)

	// newRateCounter zero period panic
	err = nil
	parl.RecoverInvocationPanic(func() {
		newRateCounter(0, counterCounters)
	}, &err)
	if err == nil {
		t.Error("RecoverInvocationPanic exp panic missing")
	} else if !strings.Contains(err.Error(), messagePeriod) {
		t.Errorf("RecoverInvocationPanic bad err: %q exp %q", err.Error(), messagePeriod)
	}
}
