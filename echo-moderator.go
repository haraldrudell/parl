/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"math"
	"time"

	"github.com/haraldrudell/parl/ptime"
)

// EchoModerator is a parallelism-limiting Moderator that:
//   - prints any increase in parallelism over concurrency
//   - prints exhibited invocation slowness exceeding latencyWarningPoint
//   - prints progressive slowness exceeding latencyWarningPoint for an
//     invocation in progress on schedule timerPeriod
//
// EchoModerator is intended to control and diagnose exec.Command invocations
type EchoModerator struct {
	moderator   ModeratorCore
	label       string
	waiting     AtomicMax[uint64]
	log         PrintfFunc
	invokeTimer InvokeTimer
}

// NewEchoModerator returns a parallelism-limiting moderator with printouts for
// excessive slowness or parallelism
func NewEchoModerator(
	concurrency uint64,
	latencyWarningPoint time.Duration,
	waitingWarningPoint uint64,
	timerPeriod time.Duration,
	label string, g0 GoGen, log PrintfFunc) (echoModerator *EchoModerator) {
	if latencyWarningPoint < minLatencyWarningPoint {
		latencyWarningPoint = minLatencyWarningPoint
	}
	if label == "" {
		label = "echoModerator"
	}
	e := EchoModerator{
		moderator: *NewModeratorCore(concurrency),
		label:     label,
		log:       log,
	}
	e.invokeTimer = *NewInvokeTimer(e.callback, latencyWarningPoint, math.MaxUint64,
		timerPeriod, g0)
	e.waiting.Value(waitingWarningPoint)
	return &e
}

func (em *EchoModerator) Do(fn func()) {

	// if highest pending request, log that
	if _, _, waiting := em.moderator.Status(); em.waiting.Value(waiting) {
		age, threadID := em.invokeTimer.Oldest()
		var threadStr string
		if threadID.IsValid() {
			threadStr = "oldest thread ID: " + threadID.String()
		}
		em.log("%s new waiting threads max: %d slowest operation: %s%s",
			em.label, waiting+1, ptime.Duration(age), threadStr)
	}

	em.moderator.Do((echoModerator{fn: fn, EchoModerator: em}).criticalSection)
}

type echoModerator struct {
	fn func()
	*EchoModerator
}

// criticalSection may have multiple threads executing
func (em echoModerator) criticalSection() {
	em.invokeTimer.Do(em.fn)
}

func (em *EchoModerator) callback(
	reason CBReason,
	maxParallelism uint64,
	maxLatency time.Duration,
	threadID ThreadID) {
	em.log("%s %s: max parallelism: %d max latency: %s goroutine-ID: %s",
		em.label, reason, maxParallelism, maxLatency, threadID)
}
