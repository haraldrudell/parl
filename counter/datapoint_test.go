/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package counter

import (
	"strings"
	"testing"
	"time"

	"github.com/haraldrudell/parl"
)

func Test_newDatapoint(t *testing.T) {
	messagePeriod := "period must be positive"
	value0 := uint64(1)
	value1 := uint64(11)
	expAverage := (value0 + value1) / 2
	period := time.Second
	now := time.Now()
	future := now.Add(period)
	tPeriod0Start := time.Now().Truncate(period)
	tPeriod1Start := tPeriod0Start.Add(period)
	tPeriod2Start := tPeriod0Start.Add(period)

	var err error
	var datapointProvider parl.Datapoint
	var datapointConsumer parl.DatapointValue
	var datapoint *Datapoint
	var datapoint2 parl.Datapoint
	var value uint64
	var max uint64
	var min uint64
	var isValid bool
	var average float64
	var n uint64

	_ = expAverage
	_ = future
	_ = tPeriod1Start
	_ = tPeriod2Start
	_ = datapoint2

	// instantiate datapoint
	datapointProvider = newDatapoint(period)
	datapointConsumer = datapointProvider.(parl.DatapointValue)
	datapoint = datapointProvider.(*Datapoint)
	if datapoint.period != period {
		t.Errorf("datapoint bad period: %s exp %s", datapoint.period, period)
	}

	// add first value
	datapointProvider.SetValue(value0)
	value, max, min, isValid, _, _ = datapointConsumer.GetDatapoint()
	if value != value0 {
		t.Errorf("datapoint1 bad value: %d exp %d", value, value0)
	}
	if max != value0 {
		t.Errorf("datapoint1 bad max: %d exp %d", max, value0)
	}
	if min != value0 {
		t.Errorf("datapoint1 bad min: %d exp %d", min, value0)
	}
	if datapoint.periodStart.IsZero() {
		t.Error("datapoint1 periodStart zero")
	}
	if !isValid {
		t.Error("datapoint1 isValid false")
	}
	if datapoint.periodStart != tPeriod0Start {
		t.Errorf("datapoint1 bad periodStart: %s exp %s",
			datapoint.periodStart.Format(parl.Rfc3339ns),
			tPeriod0Start.Format(parl.Rfc3339ns))
	}
	if datapoint.isFullPeriod {
		t.Error("datapoint1 isFullPeriod true")
	}

	// DatapointValue
	value = datapointConsumer.DatapointValue()
	if value != value0 {
		t.Errorf("DatapointValue bad value: %d exp %d", value, value0)
	}

	// DatapointMax
	max = datapointConsumer.DatapointMax()
	if max != value0 {
		t.Errorf("DatapointMax bad max: %d exp %d", max, value0)
	}

	// DatapointMin
	min = datapointConsumer.DatapointMin()
	if min != value0 {
		t.Errorf("DatapointMin bad min: %d exp %d", min, value0)
	}

	// add second value
	datapointProvider.SetValue(value1)
	value, max, min, isValid, average, n = datapointConsumer.GetDatapoint()
	if value != value1 {
		t.Errorf("datapoint2 bad value: %d exp %d", value, value1)
	}
	if max != value1 {
		t.Errorf("datapoint2 bad max: %d exp %d", max, value1)
	}
	if min != value0 {
		t.Errorf("datapoint2 bad min: %d exp %d", min, value0)
	}
	if datapoint.periodStart.IsZero() {
		t.Error("datapoint2 periodStart zero")
	}
	if !isValid {
		t.Error("datapoint2 isValid false")
	}

	// end period
	_ = average
	_ = n

	// newDatapoint zero period panic
	if err = invokeNewDatapoint(); err == nil {
		t.Error("RecoverInvocationPanic exp panic missing")
	} else if !strings.Contains(err.Error(), messagePeriod) {
		t.Errorf("RecoverInvocationPanic bad err: %q exp %q", err.Error(), messagePeriod)
	}
}

func invokeNewDatapoint() (err error) {
	defer parl.PanicToErr(&err)

	newDatapoint(0)

	return
}
