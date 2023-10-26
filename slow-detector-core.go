/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/ptime"
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

	max       AtomicMax[time.Duration]
	alwaysMax AtomicMax[time.Duration]
	last      time.Duration // atomic
	average   ptime.Averager[time.Duration]
}

func NewSlowDetectorCore(callback CbSlowDetector, slowTyp slowType, goGen GoGen, nonReturnPeriod ...time.Duration) (slowDetector *SlowDetectorCore) {
	if callback == nil {
		panic(perrors.NewPF("callback cannot be nil"))
	}

	// threshold 1: time between non-return reports, default 1 minute
	var nonReturnPeriod0 time.Duration
	if len(nonReturnPeriod) > 0 {
		nonReturnPeriod0 = nonReturnPeriod[0]
	} else {
		nonReturnPeriod0 = defaultNonReturnPeriod
	}

	return &SlowDetectorCore{
		ID:       slowIDGenerator.ID(),
		callback: callback,
		thread:   NewSlowDetectorThread(slowTyp, nonReturnPeriod0, goGen),
		max:      *NewAtomicMax[time.Duration](),
		average:  *ptime.NewAverager[time.Duration](),
	}
}

func (s *SlowDetectorCore) Init(threshold ...time.Duration) (slowDetector *SlowDetectorCore) {
	slowDetector = s
	// threshold0: minimum slowness to report, default 100 ms
	var threshold0 time.Duration
	if len(threshold) > 0 {
		threshold0 = threshold[0]
	} else {
		threshold0 = defaultMinReportDuration
	}
	s.max.Init(threshold0)
	return
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

func (sd *SlowDetectorCore) Values() (
	last, average, max time.Duration,
	hasValue bool,
) {
	last = time.Duration(atomic.LoadInt64((*int64)(&sd.last)))
	averageFloat, _ := sd.average.Average()
	average = time.Duration(averageFloat)
	max, hasValue = sd.alwaysMax.Max()
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

	// store last and average
	duration := t1.Sub(sdi.t0)
	atomic.StoreInt64((*int64)(&sd.last), int64(duration))
	sd.average.Add(duration, t1)
	sd.alwaysMax.Value(duration)

	// check against max
	if sd.max.Value(duration) {
		sd.callback(sdi, true, duration)
	}
}
