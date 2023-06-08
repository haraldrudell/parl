/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"time"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps"
	"github.com/haraldrudell/parl/ptime"
	"github.com/haraldrudell/parl/sets"
)

const (
	ITParallelism CBReason = iota + 1
	ITLatency

	minLatencyWarningPoint = 10 * time.Millisecond
	defaultTimer           = 10 * time.Second
)

// CBReason explains to consumer why the callback was invoked
//   - ITParallelism ITLatency
type CBReason uint8

// emid is a unique ordered ID for invocations usable as map key
type emid uint64

// emidGenerator.ID produces a unique echo moderator invocation ID
var emidGenerator UniqueIDTypedUint64[emid]

// CBFunc is a thread-safe function invoked on
//   - parallelism exceeding parallelismWarningPoint
//   - latency of an ongoing invocation exceeds latencyWarningPoint
type CBFunc func(reason CBReason, maxParallelism uint64, maxLatency time.Duration, threadID ThreadID)

// InvokeTimer monitors funtion invocations for parallelism and latency
//   - callback is invoked on exceeding thresholds and reaching a new max
type InvokeTimer struct {
	callback    CBFunc
	timerPeriod time.Duration
	// map of invocations with value order oldest first
	invoTimes   pmaps.ThreadSafeOrderedMapFunc[emid, *invokeTimerInvo]
	invos       AtomicCounter
	latency     AtomicMax[time.Duration]
	parallelism AtomicMax[uint64]

	goLock   sync.Mutex
	cancelGo func()

	g0 GoGen
}

// NewInvokeTimer returnds an object alerting of max latency and parallelism
//   - Do is used for new invocations
func NewInvokeTimer(
	callback CBFunc,
	latencyWarningPoint time.Duration,
	parallelismWarningPoint uint64,
	timerPeriod time.Duration, g0 GoGen) (invokeTimer *InvokeTimer) {
	if callback == nil {
		panic(perrors.NewPF("callback cannot be nil"))
	}
	if timerPeriod < defaultTimer {
		timerPeriod = defaultTimer
	}
	var ix *invokeTimerInvo
	i := InvokeTimer{
		callback:    callback,
		timerPeriod: timerPeriod,
		invoTimes:   *pmaps.NewThreadSafeOrderedMapFunc[emid](ix.oldestFirst),
		g0:          g0,
	}
	i.latency.Value(latencyWarningPoint)
	i.parallelism.Value(parallelismWarningPoint)
	return &i
}

// Do invokes fn with alerts on latency and parallelism
//   - Do is invoked in the goroutine to execute fn
func (i *InvokeTimer) Do(fn func()) {
	defer newInvokeTimerInvo(i).init().deferFunc()

	// execute
	fn()
}

func (em *InvokeTimer) Oldest() (age time.Duration, threadID ThreadID) {
	list := em.invoTimes.List(1)
	if len(list) == 0 {
		return
	}
	invocation := list[0]

	age = time.Since(invocation.t0)
	if age2, _ := em.latency.Max(); age2 > age {
		age = age2
	}
	threadID = invocation.threadID

	return
}

func (em *InvokeTimer) ensureTimer() {
	em.goLock.Lock()
	defer em.goLock.Unlock()

	em.invos.Inc()

	if em.cancelGo != nil {
		return // timer already running return
	}

	// launch timer
	subGo := em.g0.SubGo()
	g0 := subGo.Go()
	em.cancelGo = subGo.Cancel // save the cancel function for the goroutine
	go ptime.OnTimedThread(em.timerLatencyCheck, em.timerPeriod, time.Local, g0)
}

func (em *InvokeTimer) maybeCancelTimer() {
	em.goLock.Lock()
	defer em.goLock.Unlock()

	if em.invos.Dec() != 0 {
		return // more invocations are active
	} else if em.cancelGo == nil {
		return // no timer running return
	}

	// cancel timer
	em.cancelGo()
	em.cancelGo = nil
}

func (em *InvokeTimer) timerLatencyCheck(at time.Time) {

	// get age of oldest request being processed
	var age time.Duration
	var threadID ThreadID
	if list := em.invoTimes.List(1); len(list) == 0 {
		return // no requests being processed
	} else {
		invocationXp := list[0]
		age = time.Since(invocationXp.t0)
		threadID = invocationXp.threadID
	}

	// print if slowest yet
	if em.latency.Value(age) {
		max, _ := em.parallelism.Max()
		// callback for high latency of task in progress
		em.callback(ITLatency, max, age, threadID)
	}
}

func (ws CBReason) String() (s string) {
	return cbReasonSet.StringT(ws)
}

var cbReasonSet = sets.NewSet(sets.NewElements[CBReason](
	[]sets.SetElement[CBReason]{
		{ValueV: ITParallelism, Name: "max parallel"},
		{ValueV: ITLatency, Name: "slowest"},
	}))
