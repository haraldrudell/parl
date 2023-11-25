/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"math"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl/ptime"
)

const (
	// invocations of less invocation that 10 ms are not reported
	minLatencyWarningPoint = 10 * time.Millisecond
)

var echoModeratorID atomic.Uint64

type mcReturnTicket func()

// EchoModerator is a parallelism-limiting Moderator that:
//   - prints any increase in parallelism over the concurrency value
//   - prints exhibited invocation slowness exceeding latencyWarningPoint
//   - prints progressive slowness exceeding latencyWarningPoint for
//     non-returning invocations in progress on schedule timerPeriod
//   - EchoModerator can be used in production for ongoing diagnose of apis and
//     libraries
//   - cost is one thread, one timer, and a locked linked-list of invocations
//   - —
//   - EchoModerator is intended to control and diagnose exec.Command invocations
//   - problems include:
//   - — too many parallel invocations
//   - — invocations that do not return or are long running
//   - — too many threads held waiting to invoke
//   - — unexpected behavior under load
//   - — deviating behavior when operated for extended periods of time
type EchoModerator struct {
	// moderator limits parallelism
	moderator ModeratorCore
	// label preceds all printouts, default is “echoModerator1”
	label string
	// waiting causes printout if too many threads are waiting at the moderator
	waiting AtomicMax[uint64]
	log     PrintfFunc
	// examines individual invocations
	invocationTimer InvocationTimer[mcReturnTicket]
}

// NewEchoModerator returns a parallelism-limiting moderator with printouts for
// excessive slowness or parallelism
//   - concurrency is the highest number of executions that can take place in parallel
//   - printout on:
//   - — too many threads waiting at the moderator
//   - — too slow or hung invocations
//   - stores self-referencing pointers
func NewEchoModerator(
	concurrency uint64,
	latencyWarningPoint time.Duration,
	waitingWarningPoint uint64,
	timerPeriod time.Duration,
	label string, goGen GoGen, log PrintfFunc,
) (echoModerator *EchoModerator) {
	if latencyWarningPoint < minLatencyWarningPoint {
		latencyWarningPoint = minLatencyWarningPoint
	}
	if label == "" {
		label = "echoModerator" + strconv.Itoa(int(echoModeratorID.Add(1)))
	}
	m := EchoModerator{
		moderator: *NewModeratorCore(concurrency),
		label:     label,
		log:       log,
		waiting:   *NewAtomicMax(waitingWarningPoint),
	}
	m.invocationTimer = *NewInvocationTimer[mcReturnTicket](
		m.loggingCallback, m.returnMcTicket,
		latencyWarningPoint,
		// no parallelism warnings
		//	- instead warning on too many threads waiting at moderator
		math.MaxUint64,
		timerPeriod, goGen,
	)
	return &m
}

// Ticket waits for a EchoModerator ticket and provides a function to return it
//
//	func moderatedFunc() {
//	  defer echoModerator.Ticket()()
func (m *EchoModerator) Ticket() (returnTicket func()) {

	// if highest pending request, log that
	if _, _, waiting := m.moderator.Status(); m.waiting.Value(waiting) {
		age, threadID := m.invocationTimer.Oldest()
		var threadStr string
		if threadID.IsValid() {
			threadStr = "oldest thread ID: " + threadID.String()
		}
		m.log("%s new waiting threads max: %d slowest operation: %s%s",
			m.label, waiting+1, ptime.Duration(age), threadStr)
	}

	// blocks here
	var ticketReturn mcReturnTicket = m.moderator.Ticket()

	// hand the ticket return to invocation
	//	- to avoid additional object creation, invocation
	//		will safekeep the ticket return and provide it
	//		via callback at end of invocation
	//	- invocation will invoke returnMcTicket with it
	returnTicket = m.invocationTimer.Invocation(ticketReturn)
	return
}

// returnMcTicket receives tickets to be returned from an ending Invocation
func (m *EchoModerator) returnMcTicket(ticketReturn mcReturnTicket) {
	ticketReturn()
}

// loggingCallback logs output from invocationTimer
//   - there is also logging in Ticket
func (m *EchoModerator) loggingCallback(
	reason CBReason,
	maxParallelism uint64,
	maxLatency time.Duration,
	threadID ThreadID) {
	m.log("%s %s: max parallelism: %d max latency: %s goroutine-ID: %s",
		m.label, reason, maxParallelism, maxLatency, threadID)
}
