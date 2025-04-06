/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"time"

	"github.com/haraldrudell/parl/ptime"
)

// SlowDetectorCore measures latency via Start-Stop invocations
//   - a thread measures time of non-returning, hung invocations
type SlowDetectorCore struct {
	// slowID is unique opaque identifier [constraints.Ordered] typically integral
	//	- used as map key
	ID slowID
	// reportReceiver receives reports for the slowest-to-date invocation
	// and non-return reports every minute
	reportReceiver SlowReporter
	// thread watches non-returning invocations
	//	- may be a shared object so must be pointer
	thread *SlowDetectorThread
	// endr is provided to [NewSlowDetectorInvocation]
	//	- it provides client methods Duration Report Stop
	//	- it is privatee struct delegating those methods to this SlowDetectorCore
	//	- implements [SlowDetectorIf] SlowDetectorInvocationEnder
	endr ender

	// thread-safe field after new-function

	// max holds the minimum value to produce slow-invocation report
	max AtomicMax[time.Duration]
	// alwaysMax is the slowest ever invocation
	alwaysMax AtomicMax[time.Duration]
	// last is latency for last invocation
	last Atomic64[time.Duration]
	// average is average latency across all invocations
	average ptime.Averager[time.Duration]
}

// NewSlowDetectorCore returns an object tracking non-returning or slow function invocations
//   - reportReceiver: receives offending slow-detector invocations [SlowReporter.Report]
//   - slowTyp SlowDefault: shared thread: recommended
//   - — the thread scans a map for non-returning invocations
//   - slowTyp SlowOwnThread: dedicated thread that never exits
//   - slowTyp SlowShutdownThread: dedicated thread that exits whenever all invocations ends
//   - goGen: used to start shared or dedicated thread
//   - nonReturnPeriod is one or two timing values:
//   - — nonReturnPeriod: how often non-returning invocations are reported, default once per minute
//   - — minReportedDuration: minimum slowness duration that is being reported, default 100 ms
func NewSlowDetectorCore(
	fieldp *SlowDetectorCore, reportReceiver SlowReporter, slowTyp SlowType,
	goGen GoGen, nonReturnPeriod ...time.Duration,
) (slowDetector *SlowDetectorCore) {
	NilPanic("reportReceiver", reportReceiver)

	// get slowDetector
	if fieldp != nil {
		slowDetector = fieldp
	} else {
		slowDetector = &SlowDetectorCore{}
	}

	// nonReturnPeriod[0]: time between non-return reports, default 1 minute
	var nonReturnPeriod0 time.Duration
	if len(nonReturnPeriod) > 0 {
		nonReturnPeriod0 = nonReturnPeriod[0]
	} else {
		nonReturnPeriod0 = defaultNonReturnPeriod
	}

	// nonReturnPeriod[1]: minimum duration for slowness to be reported, default 100 ms
	var minReportedDuration time.Duration
	if len(nonReturnPeriod) > 1 {
		minReportedDuration = nonReturnPeriod[1]
	} else {
		minReportedDuration = defaultMinReportDuration
	}

	*slowDetector = SlowDetectorCore{
		ID:             slowIDGenerator.ID(),
		reportReceiver: reportReceiver,
		thread:         NewSlowDetectorThread(slowTyp, nonReturnPeriod0, goGen),
		average:        *ptime.NewAverager[time.Duration](),
	}
	NewAtomicMaxp(&slowDetector.max, minReportedDuration)
	slowDetector.endr.sdc = slowDetector
	return
}

// Start returns the effective start time for a new timing cycle
//   - invoLabel: printable identifier for the invocation, often short code location:
//     “mains.(*Executable).AddErr-executable.go:25”
//   - timeStamp: optional start time, default time.Now()
func (s *SlowDetectorCore) Start(invoLabel string, timeStamp ...time.Time) (invocation *SlowDetectorInvocation) {

	// get time value for this operation
	var t0 time.Time
	if len(timeStamp) > 0 {
		t0 = timeStamp[0]
	} else {
		t0 = time.Now()
	}

	// save in map, launch thread if not already running
	//	- thread stores invocation in map, so it must be on heap
	invocation = NewSlowDetectorInvocation(slowIDGenerator.ID(), invoLabel, goID(), t0, &s.endr)
	s.thread.Start(invocation)

	return
}

// Values returns statistics metrics
//   - last: latency of last invocation, zero when none
//   - average: average latency for all invocations, zero when none
//   - max: the slowest invocation if hasValue true
//   - hasValue: indicates if values are valid, false for no invocations
func (s *SlowDetectorCore) Values() (
	last, average, max time.Duration,
	hasValue bool,
) {
	last = s.last.Load()
	var averageFloat, _ = s.average.Average()
	average = time.Duration(averageFloat)
	max, hasValue = s.alwaysMax.Max()
	return
}

// reportDuration returns true if this is a progressive max
//   - isNewMax true: this invocation is a new progressive max for this slow-detector
//   - also records all-time max
func (s *SlowDetectorCore) reportDuration(duration time.Duration) (isNewMax bool) {
	s.alwaysMax.Value(duration)
	isNewMax = s.max.Value(duration)
	return
}

// stop handles the end of an invocation
//   - invocation: the invocation that ended
//   - timestamp: the ending time, default now
//   - stop is delegated from [SlowDetectorInvocation.Stop]
func (s *SlowDetectorCore) stop(invocation *SlowDetectorInvocation, timestamp ...time.Time) {

	// remove invoaction from map and possibly shutdown thread
	s.thread.Stop(invocation)

	// get time value for this operation
	var t1 time.Time
	if len(timestamp) > 0 {
		t1 = timestamp[0]
	} else {
		t1 = time.Now()
	}

	// store last and average

	// duration is elapsed time for this invocation
	var duration = t1.Sub(invocation.t0)
	s.last.Store(duration)
	s.average.Add(duration, t1)
	s.alwaysMax.Value(duration)

	// check against max
	if s.max.Value(duration) {
		s.reportReceiver.Report(invocation, DidReturnYes, duration)
	}
}

const (
	// minReportedDuration: minimum slowness duration that is
	// being reported, default 100 ms
	defaultMinReportDuration = 100 * time.Millisecond
	// default how often non-returning invocations are reported,
	// once per minute
	defaultNonReturnPeriod = time.Minute
)

// slowID is a unique identifier for slow-detector entities
//   - usable as map key
type slowID uint64

// slowIDGenerator generates unique ID values
var slowIDGenerator UniqueIDTypedUint64[slowID]

// ender is SlowDetectorInvocationEnder
type ender struct{ sdc *SlowDetectorCore }

// ender is SlowDetectorInvocationEnder
var _ SlowDetectorInvoActionsStop = &ender{}

// Stop ends an invocation created by SlowDetectorCore
//   - invocation: th einvocation object
//   - timestamp: optional ending timestamp, default now
func (e *ender) Stop(invocation *SlowDetectorInvocation, timestamp ...time.Time) {
	e.sdc.stop(invocation, timestamp...)
}

// Duration returns true if this is a progressive max
//   - isNewMax true: this invocation is a new progressive max for this slow-detector
//   - also records all-time max
func (e *ender) Duration(duration time.Duration) (isNewMax bool) {
	return e.sdc.reportDuration(duration)
}

// Report receives reports for the slowest-to-date invocation
// and non-return reports every minute
//   - invocation: the invocation created by [SlowDetectorCode.Start]
//   - didReturn true: the invocation has ended
//   - didReturn false: the invocation is still in progress, ie. a non-return report
//   - duration: the latency causing report
func (e *ender) Report(invocation *SlowDetectorInvocation, didReturn DidReturn, duration time.Duration) {
	e.sdc.reportReceiver.Report(invocation, didReturn, duration)
}
