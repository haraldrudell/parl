/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"time"

	"github.com/haraldrudell/parl/perrors"
)

const (
	defaultMinReportDuration = 100 * time.Millisecond
	defaultNonReturnPeriod   = time.Minute
)

type CbSlowDetector func(sdi *SlowDetectorInvocation, hasReturned bool, duration time.Duration)
type slowID uint64

var slowIDGenerator UniqueIDTypedUint64[slowID]

// SlowDetectorCore measures latency via Start-Stop invocations
//   - Thread-Safe and multi-threaded, parallel invocations
//   - Separate thread measures time of non-returning, hung invocations
type SlowDetectorCore struct {
	ID       slowID
	callback CbSlowDetector
	thread   *SlowDetectorThread
	max      AtomicMax[time.Duration]
}

func NewSlowDetectorCore(callback CbSlowDetector, slowTyp slowType, goGen GoGen, threshold ...time.Duration) (slowDetector *SlowDetectorCore) {
	if callback == nil {
		panic(perrors.NewPF("callback cannot be nil"))
	}

	// threshold0: minimum slowness to report, default 100 ms
	var threshold0 time.Duration
	if len(threshold) > 0 {
		threshold0 = threshold[0]
	} else {
		threshold0 = defaultMinReportDuration
	}

	// threshold 1: time between non-return reports, default 1 minute
	var nonReturnPeriod time.Duration
	if len(threshold) > 1 {
		nonReturnPeriod = threshold[1]
	} else {
		nonReturnPeriod = defaultNonReturnPeriod
	}

	return &SlowDetectorCore{
		ID:       slowIDGenerator.ID(),
		callback: callback,
		thread:   NewSlowDetectorThread(slowTyp, nonReturnPeriod, goGen),
		max:      *NewAtomicMax(threshold0),
	}
}

// Start returns the effective start time for a new timing cycle
//   - value is optional start time, default time.Now()
func (sd *SlowDetectorCore) Start(invoLabel string, value ...time.Time) (invocation *SlowDetectorInvocation) {

	// get time value for this operation
	var t0 time.Time
	if len(value) > 0 {
		t0 = value[0]
	} else {
		t0 = time.Now()
	}

	// save in map, launch thread if not already running
	s := SlowDetectorInvocation{
		sID:       slowIDGenerator.ID(),
		invoLabel: invoLabel,
		threadID:  goID(),
		t0:        t0,
		stop:      sd.stop,
		sd:        sd,
	}
	sd.thread.Start(&s)
	return &s
}

func (sd *SlowDetectorCore) Max() (max time.Duration, hasValue bool) {
	max, hasValue = sd.max.Max()
	return
}

// Stop is invoked via SlowDetectorInvocation
func (sd *SlowDetectorCore) stop(sdi *SlowDetectorInvocation, value ...time.Time) {

	// remove from map and possibly shutdown thread
	sd.thread.Stop(sdi)

	// get time value for this operation
	var t1 time.Time
	if len(value) > 0 {
		t1 = value[0]
	} else {
		t1 = time.Now()
	}

	// check against max
	if duration := t1.Sub(sdi.t0); sd.max.Value(duration) {
		sd.callback(sdi, true, duration)
	}
}