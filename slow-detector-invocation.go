/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"strconv"
	"strings"
	"time"

	"github.com/haraldrudell/parl/ptime"
)

// SlowDetectorInvocation is a container used by SlowDetectorCore
type SlowDetectorInvocation struct {
	// slowID is unique opaque identifier [constraints.Ordered] typically integral
	//	- used as map key
	sID slowID
	// threadID is the Go goroutine identifier integer that
	// invoked [SlowDetectorCore.Start] creating this invocation
	//	- typically a numeric string
	threadID ThreadID
	// invoLabel is printable identifier for the invocation, often short code location:
	//	- “build”
	//	- “mains.(*Executable).AddErr-executable.go:25”
	invoLabel string
	// t0 is the timestamp for when the invocation started,
	// default [time.Now] of [SlowDetectorCore.Start] invocation
	t0 time.Time
	// ender handles [SlowDetectorInvocation.Stop] invocations
	//	- Duration() Report() Stop()
	ender SlowDetectorInvoActionsStop
	// timestamp for last non-return report.
	//	- [ptime.Epoch] fits timestamp into integral atomic
	lastNonReturnReportTimestamp Atomic64[ptime.Epoch]
	// intervalsLock makes intervals slice thread-safe
	intervalsLock Mutex
	// intervals is labeled progress timestamps during
	// the life of an invocation
	//	- accessed behind lock
	intervals []interval
}

// SlowDetectorInvocation is SlowInvocation
var _ SlowInvocation = &SlowDetectorInvocation{}

// NewSlowDetectorInvocation retruns a data container and
// invocation-action provider for an invocation
//   - slowID: unique opaque identifier [constraints.Ordered] typically integral
//   - — used as map key
//   - invoLabel: printable identifier for the invocation, often short code location:
//   - — “build”
//   - — “mains.(*Executable).AddErr-executable.go:25”
//   - threadID: Go goroutine identifier created this invocation
//   - t0: the timestamp for when the invocation started,
//     default [time.Now] of [SlowDetectorCore.Start] invocation
//   - ender: provides invocation actions: Duration() Report() Stop()
//   - —
//   - heap allocation to be put in maps
func NewSlowDetectorInvocation(ID slowID, invoLabel string, threadID ThreadID, t0 time.Time, ender SlowDetectorInvoActionsStop) (invocation *SlowDetectorInvocation) {
	return &SlowDetectorInvocation{
		sID:       ID,
		invoLabel: invoLabel,
		threadID:  threadID,
		t0:        t0,
		ender:     ender,
	}
}

// Stop ends an invocation created by SlowDetectorCore
func (s *SlowDetectorInvocation) Stop(value ...time.Time) { s.ender.Stop(s, value...) }

// Interval adds a labelled timestamp to an ongoing invocation
//   - label: printable timestamp identifier “lsofComplete”
//   - t: optional timestamp, default now
func (s *SlowDetectorInvocation) Interval(label string, t ...time.Time) {

	// timestamp default now
	var t0 time.Time
	if len(t) > 0 {
		t0 = t[0]
	}
	if t0.IsZero() {
		t0 = time.Now()
	}
	defer s.intervalsLock.Lock().Unlock()

	// append label and timestamp to intervals
	if label == "" {
		label = strconv.Itoa(len(s.intervals) + 1)
	}
	s.intervals = append(s.intervals, interval{label: label, t: t0})
}

// ThreadID returns the thread ID dor the thread invoking Start
func (s *SlowDetectorInvocation) ThreadID() (threadID ThreadID) { return s.threadID }

// T0 returns Start invocation timestamp
func (s *SlowDetectorInvocation) T0() (t0 time.Time) { return s.t0 }

// Label returns the invocation label
func (s *SlowDetectorInvocation) Label() (label string) { return s.invoLabel }

// Time returns the last non-return report timestamp
//   - t zero-value: retrieve any previous timestamp
//   - t: new timestamp to set
//   - previousT: retrieved or previous timestamp, zero-time for none
func (s *SlowDetectorInvocation) Time(t time.Time) (previousT time.Time) {
	if t.IsZero() {
		// t zero-value means retrieve current value
		previousT = s.lastNonReturnReportTimestamp.Load().Time()
	} else {
		previousT = s.lastNonReturnReportTimestamp.Swap(ptime.EpochNow(t)).Time()
	}
	return
}

// Intervals returns printable space-separated string of intervals
//   - printable label “lsofComplete” and a time relative to initial Start
func (s *SlowDetectorInvocation) Intervals() (intervalStr string) {
	defer s.intervalsLock.Lock().Unlock()

	if len(s.intervals) > 0 {
		var sList = make([]string, len(s.intervals))
		var t0 = s.t0
		for i, ivl := range s.intervals {
			var t = ivl.t
			sList[i] = ptime.Duration(t.Sub(t0)) + "\x20" + ivl.label
			t0 = t
		}
		intervalStr = "\x20" + strings.Join(sList, "\x20")
	}

	return
}

// InvoActions returns an action object for this invocation
//   - Duration() Report() Stop()
func (s *SlowDetectorInvocation) InvoActions() (sdcIf SlowDetectorInvoActions) {
	return s.ender
}

// interval is a labelled timestamp for an ongoing invocation
type interval struct {
	label string
	t     time.Time
}
