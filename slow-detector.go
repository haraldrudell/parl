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

const (
	newFrames    = 1
	ensureFrames = 1
)

// SlowDetector measures latency via Start-Stop invocations and prints
// max latency values above threshold to stderr. Thread-safe
type SlowDetector struct {
	sd     SlowDetectorCore
	printf PrintfFunc
	label  string
}

type SlowInvocation interface {
	Stop(value ...time.Time)
	Interval(label string, t ...time.Time)
}

// NewSlowDetector returns a Start-Stop variable detecting slowness
//   - label is a name for the measured activity, default the code location of caller
//   - slowTyp is most commonly SlowDefault using a shared thread
//   - default printf is parl.Log, ie. ouput to stderr
//   - first optional duration is minimum latency to report, default 100 ms
//     if first optional duration is 0, all max-slowness invocations are printed
//   - second optional duration is reporting period of non-return, default 1 minute
//   - output is to stderr
func NewSlowDetector(label string, slowTyp slowType, printf PrintfFunc, goGen GoGen, threshold ...time.Duration) (slowDetector *SlowDetector) {
	if label == "" {
		label = pruntime.NewCodeLocation(newFrames).Short()
	}
	if printf == nil {
		printf = Log
	}
	sd := SlowDetector{
		label:  label,
		printf: printf,
	}
	var threshold0, nonReturnPeriod []time.Duration
	if len(threshold) > 0 {
		threshold0 = threshold[0:1]
		if len(threshold) > 1 {
			nonReturnPeriod = threshold[1:2]
		}
	}
	sd.sd = *NewSlowDetectorCore(sd.callback, slowTyp, goGen, nonReturnPeriod...)
	sd.sd.Init(threshold0...)
	return &sd
}

func (sd *SlowDetector) IsValid() (isValid bool) {
	if sd == nil {
		return
	}
	isValid = sd.printf != nil
	return
}

func (sd *SlowDetector) Start0() (slowInvocation SlowInvocation) {
	return sd.sd.Start(pruntime.NewCodeLocation(ensureFrames).Short())
}

func (sd *SlowDetector) Start(label string, value ...time.Time) (slowInvocation SlowInvocation) {
	if label == "" {
		label = pruntime.NewCodeLocation(ensureFrames).Short()
	}
	return sd.sd.Start(label, value...)
}

func (sd *SlowDetector) Values() (last, average, max time.Duration, hasValue bool) {
	return sd.sd.Values()
}

// last-duration / average duration / max duration
func (sd *SlowDetector) Status0() (s string) {
	last, average, max, hasValue := sd.sd.Values()
	if !hasValue {
		return "-/-/-"
	}
	s = ptime.Duration(last) + "/" +
		ptime.Duration(average) + "/" +
		ptime.Duration(max)
	return
}

func (sd *SlowDetector) Status() (s string) {
	return sd.label + ": " + sd.Status0()
}

func (sd *SlowDetector) callback(sdi *SlowDetectorInvocation, didReturn bool, duration time.Duration) {

	var inProgressStr string
	if !didReturn {
		inProgressStr = " in progress…"
	}

	var threadIDStr string
	if threadID := sdi.ThreadID(); threadID.IsValid() {
		threadIDStr = " threadID: " + threadID.String()
	}

	var intervalStr = sdi.Intervals()

	sd.printf("Slowness: %s %s duration: %s%s%s%s",
		sd.label, sdi.Label(),
		ptime.Duration(duration),
		intervalStr,
		threadIDStr,
		inProgressStr,
	)
}
