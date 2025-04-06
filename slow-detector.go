/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"time"

	"github.com/haraldrudell/parl/pruntime"
	"github.com/haraldrudell/parl/ptime"
)

// SlowDetector measures latency via Start-Stop invocations and prints
// max latency values above threshold to stderr
type SlowDetector struct {
	// sdc is the SlowDetectorCore providining mechanis to
	// monitoring invocations
	sdc SlowDetectorCore
	// log prints slowness report one-liners
	//	- “Slowness: ReBuild Rebuild duration: 101ms threadID: 212”
	log PrintfFunc
	// label is name for the measured activity
	//	- “build”
	//	- default: code location of caller:
	//		“mains.(*Executable).AddErr-executable.go:25”
	label string
}

// NewSlowDetector returns a set-and-forget invocation monitor
//   - typically used to monitor system commands that may be slow or hang
//   - label: optional activity name experiencing slownewss,
//     default the code location of caller
//   - —“build” “worker” “lsof”
//   - — default: “mains.(*Executable).AddErr-executable.go:25”
//   - slowTyp SlowDefault: shared thread: recommended
//   - — the thread scans a map for non-returning invocations
//   - slowTyp SlowOwnThread: dedicated thread that never exits
//   - — wasteful one-thread per instance
//   - slowTyp SlowShutdownThread: dedicated thread that exits whenever all invocations ends
//   - — questionable performance
//   - log: output for slowness reports
//   - log nil: default printf is [parl.Log], ie. ouput to standard error
//   - goGen: used to start shared or dedicated thread
//   - threshold: one or two optional timing values:
//   - — nonReturnPeriod: how often non-returning invocations are reported, default once per minute
//   - — minReportedDuration: minimum slowness duration that is being reported, default 100 ms
//   - — zero: all max-slowness invocations are printed.
//     May be useful to examine how slow an invocation becomes
//   - —
//   - [SlowDetector.Start0] creates invocation labeled by code location and timestamp now
//   - [SlowDetector.Start] creates labeled invocation for specific timestamp
//   - outputs one-liner reports, default to standard error
func NewSlowDetector(fieldp *SlowDetector, label string, slowTyp SlowType, log PrintfFunc, goGen GoGen, threshold ...time.Duration) (slowDetector *SlowDetector) {

	// get slowDetector
	if fieldp != nil {
		slowDetector = fieldp
	} else {
		slowDetector = &SlowDetector{}
	}

	if label == "" {
		// “mains.(*Executable).AddErr-executable.go:25”
		label = pruntime.NewCodeLocation(newFrames).Short()
	}
	if log == nil {
		log = Log
	}
	*slowDetector = SlowDetector{
		label: label,
		log:   log,
	}
	NewSlowDetectorCore(&slowDetector.sdc, slowDetector, slowTyp, goGen, threshold...)
	return
}

// IsValid returns true if SlowDetector was initialized. Thread-safe
func (s *SlowDetector) IsValid() (isValid bool) {
	if s == nil {
		return
	}
	isValid = s.log != nil
	return
}

// Start0 creates a labeled invocation labeled by code location and timestamp now
//   - label: “mains.(*Executable).AddErr-executable.go:25”
//   - Thread-safe
func (s *SlowDetector) Start0() (slowInvocation SlowInvocation) {
	// “mains.(*Executable).AddErr-executable.go:25”
	var label = pruntime.NewCodeLocation(ensureFrames).Short()
	slowInvocation = s.sdc.Start(label)
	return
}

// Start creates invocation instance
//   - label: printable label, default code location of invoker
//   - — “mains.(*Executable).AddErr-executable.go:25”
//   - timestamp: time to use, default now
//   - —
//   - heap allocation to put in maps
//   - Thread-safe
func (s *SlowDetector) Start(label string, timestamp ...time.Time) (slowInvocation SlowInvocation) {
	if label == "" {
		label = pruntime.NewCodeLocation(ensureFrames).Short()
	}
	slowInvocation = s.sdc.Start(label, timestamp...)

	return
}

// Values returns statistics metrics
//   - last: latency of last invocation, zero when none
//   - average: average latency for all invocations, zero when none
//   - max: the slowest invocation, valid when hasValue true
//   - hasValue: indicates if values are valid, false for no invocations
//   - Thread-safe
func (s *SlowDetector) Values() (last, average, max time.Duration, hasValue bool) {
	return s.sdc.Values()
}

// Status0 returns unlabeled timing results string “51ms/0s/73m” or “-/-/-”
//   - last-duration / average duration / max duration
//   - averaged 10 invocations last 10 seconds
//   - Thread-safe
func (s *SlowDetector) Status0() (s2 string) {
	var last, average, max, hasValue = s.sdc.Values()
	if !hasValue {
		return "-/-/-"
	}
	s2 = ptime.Duration(last) + "/" +
		ptime.Duration(average) + "/" +
		ptime.Duration(max)
	return
}

// Status returns labeled status “build: 1ms/2ms/3ms”
//   - last-duration / average duration / max duration
//   - averaged 10 invocations last 10 seconds
//   - Thread-safe
func (s *SlowDetector) Status() (s2 string) { return s.label + ": " + s.Status0() }

// Report receives reports for the slowest-to-date invocation
// and non-return reports every minute
//   - invo: the invocation created by [SlowDetectorCode.Start]
//   - didReturn DidReturnYes: the invocation has ended
//   - didReturn DidReturnNo: the invocation is still in progress, ie. a non-return report
//   - duration: the as-of-now latency causing report
func (s *SlowDetector) Report(invo *SlowDetectorInvocation, didReturn DidReturn, duration time.Duration) {

	// if invocation has yet to return:
	// “ in progress…”
	var inProgressStr string
	if didReturn != DidReturnYes {
		inProgressStr = " in progress…"
	}

	// threadIDStr is any thread ID as numeric string
	//	- “  threadID: 1”
	var threadIDStr string
	if threadID := invo.ThreadID(); threadID.IsValid() {
		threadIDStr = " threadID: " + threadID.String()
	}

	// space-separated string of labeled intermediate timestamps
	var intervalStr = invo.Intervals()

	s.log("Slowness: %s %s duration: %s%s%s%s",
		s.label, invo.Label(),
		ptime.Duration(duration),
		intervalStr,
		threadIDStr,
		inProgressStr,
	)
}

const (
	// [NewSlowDetector] frames
	newFrames = 1
	// [SlowDetector.Start0] frames
	ensureFrames = 1
)
