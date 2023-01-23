/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"time"

	"github.com/haraldrudell/parl/perrors"
)

// AtomicMaxDuration calculates durations maintaining max duration value
//   - Thread-Safe but designed for single thread
//   - for re-entrant timer, use SlowDetector
type AtomicMaxDuration struct {
	t0Reference AtomicReference[time.Time]
	dMax        AtomicMax[time.Duration]
}

// Start returns the effective start time for a new timing cycle
//   - value is optional start time, default time.Now()
func (ad *AtomicMaxDuration) Start(value ...time.Time) (tStart time.Time) {

	// deteremine tStart and store in atomic reference
	var previousReference *time.Time
	if tStart, previousReference = ad.do(false, value...); previousReference != nil {
		panic(perrors.ErrorfPF("two Start without Stop: %s", (*previousReference).Format(Rfc3339ns)))
	}

	return
}

// Stop returns the duration of a timing cycle
func (ad *AtomicMaxDuration) Stop(value ...time.Time) (duration time.Duration, isNewMax bool) {

	// determine tStop, retrieve atomic reference abnd calculate duration
	if tStop, previousReference := ad.do(true, value...); previousReference == nil {
		panic(perrors.ErrorfPF("Stop without Start: %s", tStop.Format(Rfc3339ns)))
	} else {
		duration = tStop.Sub(*previousReference)
	}

	// calculate maximum duration
	isNewMax = ad.dMax.Value(duration)

	return
}

// Stop returns the duration of a timing cycle
func (ad *AtomicMaxDuration) Max() (max time.Duration, hasValue bool) {
	max, hasValue = ad.dMax.Max()
	return
}

// do returns the previous reference and the active time for a Start or Stop operation
func (ad *AtomicMaxDuration) do(isStop bool, value ...time.Time) (activeTime time.Time, previousReference *time.Time) {

	// get time value for this operation
	if len(value) > 0 {
		activeTime = value[0]
	} else {
		activeTime = time.Now()
	}

	// do reference swap
	var newReference *time.Time
	if !isStop {
		newReference = &activeTime // start time to store for Start operation
	}
	previousReference = ad.t0Reference.Put(newReference)

	return
}
