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
	// SlowDetectorCore runs a thread monitoring invocations
	sdc SlowDetectorCore
	// printf prrints slowness report one-liners
	printf PrintfFunc
	// label is name for the measured activity, default the code location of caller: “build”
	label string
}

// NewSlowDetector returns a set-and-forget invocation monitor
//   - label: the activity experienceing slownewss, default the code location of caller
//   - —“build” “worker” “lsof”
//   - — “mains.(*Executable).AddErr-executable.go:25”
//   - slowTyp SlowDefault: shared thread
//   - slowTyp SlowOwnThread: dedicated thread that never exits
//   - slowTyp SlowShutdownThread: dedicated thread that exits whenever all invocations ends
//   - printf: prints slowness reports
//   - printf nil: default printf is [parl.Log], ie. ouput to standard error
//   - goGen: used to start shared or dedicated thread
//   - threshold:
//   - —
//   - [SlowDetector.Start0] creates invocation labeled by code location and timestamp now
//   - [SlowDetector.Start] creates labaled invocationfor specific timestamp
//   - output reports, default to standard error
//   - first optional duration is minimum latency to report, default 100 ms
//     if first optional duration is 0, all max-slowness invocations are printed
//   - second optional duration is reporting period of non-return, default 1 minute
func NewSlowDetector(fieldp *SlowDetector, label string, slowTyp SlowType, printf PrintfFunc, goGen GoGen, threshold ...time.Duration) (slowDetector *SlowDetector) {

	if fieldp != nil {
		slowDetector = fieldp
	} else {
		slowDetector = &SlowDetector{}
	}

	if label == "" {
		label = pruntime.NewCodeLocation(newFrames).Short()
	}
	if printf == nil {
		printf = Log
	}
	*slowDetector = SlowDetector{
		label:  label,
		printf: printf,
	}
	NewSlowDetectorCore(&slowDetector.sdc, slowDetector, slowTyp, goGen, threshold...)
	return
}

// IsValid returns true if SlowDetector was initialized. Thread-safe
func (s *SlowDetector) IsValid() (isValid bool) {
	if s == nil {
		return
	}
	isValid = s.printf != nil
	return
}

// Start0 creates invocation labeled by code location and timestamp now
//   - “mains.(*Executable).AddErr-executable.go:25”
//   - Thread-safe
func (s *SlowDetector) Start0() (slowInvocation SlowInvocation) {
	// “mains.(*Executable).AddErr-executable.go:25”
	var label = pruntime.NewCodeLocation(ensureFrames).Short()
	slowInvocation = s.sdc.Start(label)
	return
}

// Start creates invocation
//   - label: printable label, default code location of invoker
//   - — “mains.(*Executable).AddErr-executable.go:25”
//   - timestamp: time to use,defaultnow
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
//   - max: the slowest invocation if hasValue true
//   - hasValue: indicates if values are valid, false for no invocations
//   - Thread-safe
func (s *SlowDetector) Values() (last, average, max time.Duration, hasValue bool) {
	return s.sdc.Values()
}

// last-duration / average duration / max duration
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

// labeled status “build: 1ms/2ms/3ms”
//   - Thread-safe
func (s *SlowDetector) Status() (s2 string) { return s.label + ": " + s.Status0() }

func (s *SlowDetector) Report(sdi *SlowDetectorInvocation, didReturn bool, duration time.Duration) {

	var inProgressStr string
	if !didReturn {
		inProgressStr = " in progress…"
	}

	var threadIDStr string
	if threadID := sdi.ThreadID(); threadID.IsValid() {
		threadIDStr = " threadID: " + threadID.String()
	}

	var intervalStr = sdi.Intervals()

	s.printf("Slowness: %s %s duration: %s%s%s%s",
		s.label, sdi.Label(),
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
