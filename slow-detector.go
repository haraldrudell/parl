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
	sd.sd = *NewSlowDetectorCore(sd.callback, slowTyp, goGen, threshold...)
	return &sd
}

func (sd *SlowDetector) Start0() (slowInvocation SlowInvocation) {
	return sd.Start("")
}

func (sd *SlowDetector) Start(label string, value ...time.Time) (slowInvocation SlowInvocation) {
	if label == "" {
		label = pruntime.NewCodeLocation(ensureFrames).Short()
	}
	return sd.sd.Start(label, value...)
}

func (sd *SlowDetector) Max() (max time.Duration, hasValue bool) {
	max, hasValue = sd.sd.Max()
	return
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

	sd.printf("Slowness: %s duration: %s%s%s", sdi.Label(), ptime.Duration(duration),
		threadIDStr,
		inProgressStr)
}
