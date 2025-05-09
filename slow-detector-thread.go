/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl/parli"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps"
)

// SlowDetectorThread is a thread that monitors invocations for non-return
type SlowDetectorThread struct {
	// thread type identifier: [SlowDefault] [SlowOwnThread] [SlowShutdownThread]
	//	- default is thread shared across multiple slow detectors
	slowTyp SlowType
	// nonReturnPeriod is time after reporting as non-return,
	// common default 1 minute
	nonReturnPeriod time.Duration
	// active invocations being monitored
	//	- key: unique ID
	//	- value: an invocation that may fail to return
	invocations pmaps.RWMap[slowID, *SlowDetectorInvocation]
	// hasThread is lazy thread creation
	hasThread atomic.Bool
	// goGen is used for creating the thread: shared or dedicated
	goGen GoGen
	// slowLock makes subGo thread-safe and
	// creates critical section for initializing the shared thread
	slowLock Mutex
	// subGo for any running thread
	subGo SubGo
}

// NewSlowDetectorThread
//   - slowType SlowDefault: use shared thread
//   - slowType SlowOwnThread: use dedicated thread
//   - nonReturnPeriod: time after reporting as non-return, common default 1 minute
//   - goGen: used for creating the thread: shared or dedicated
//   - must be pointer because a shared value may be returned
func NewSlowDetectorThread(slowTyp SlowType, nonReturnPeriod time.Duration, goGen GoGen) (sdt *SlowDetectorThread) {
	NilPanic("goGen", goGen)

	// dedicated thread case
	if slowTyp != SlowDefault {
		sdt = &SlowDetectorThread{
			slowTyp:         slowTyp,
			nonReturnPeriod: nonReturnPeriod,
			goGen:           goGen,
		}
		pmaps.NewRWMap2(&sdt.invocations)
		return
	}
	// it is shared thread

	// sdt is shared instance
	sdt = &slowDetectorThread
	// critical section shared instance initialization
	defer sdt.slowLock.Lock().Unlock()

	if sdt.goGen != nil {
		return // slowDetectorThread already initialized return
	}

	// slowDetectorThread initialization
	sdt.slowTyp = slowTyp
	sdt.nonReturnPeriod = nonReturnPeriod
	pmaps.NewRWMap2(&sdt.invocations)
	sdt.goGen = goGen

	return
}

// Start adds an invocation to monitoring
func (s *SlowDetectorThread) Start(sdi *SlowDetectorInvocation) {

	// store in map
	s.invocations.Put(sdi.sID, sdi)

	if s.hasThread.Load() || !s.hasThread.CompareAndSwap(false, true) {
		return // thread already running return
	}

	// launch thread
	var g = s.newSubGo()
	if g == nil {
		return
	}
	go s.thread(g)
}

// Stop removes an invocation from being monitored
func (s *SlowDetectorThread) Stop(invo *SlowDetectorInvocation) {

	// remove from map
	s.invocations.Delete(invo.sID, parli.MapDeleteWithZeroValue)

	// check wheether thread may be shut down
	if s.slowTyp != SlowShutdownThread || s.invocations.Length() > 0 {
		return // not to be shutdown or not to be shutdown now return
	}

	// exit the thread
	defer s.slowLock.Lock().Unlock()

	var subGo = s.subGo
	if subGo == nil {
		panic(perrors.NewPF("spurios SlowDetectorThread.Stop"))
	}
	s.subGo = nil
	s.hasThread.Store(false)

	s.subGo.Cancel()
}

// newSubGo creeates new subGo
//   - g non-nil: do start the thread
func (s *SlowDetectorThread) newSubGo() (g Go) {
	defer s.slowLock.Lock().Unlock()

	if s.invocations.Length() == 0 {
		return // no invocations
	}
	s.subGo = s.goGen.SubGo()
	g = s.subGo.Go()

	return
}

// thread until context cancel or Stop of last invocation
func (s *SlowDetectorThread) thread(g Go) {
	var err error
	defer g.Register("SlowDetectorThread" + goID().String()).Done(&err)
	defer RecoverErr(func() DA { return A() }, &err)

	// ticker starts scan for non-returns every second
	var ticker = time.NewTicker(slowScanPeriod)
	defer ticker.Stop()

	var C <-chan time.Time = ticker.C
	var done <-chan struct{} = g.Context().Done()
	var t time.Time
	for {
		select {
		case <-done:
			return // context cancel return
		case <-C:
			t = time.Now()
		}
		// ticker triggered scan of non-returns

		// check all invocations for non-return
		for _, invocation := range s.invocations.List() {

			// duration is how long the invocation has been in progress
			var duration = t.Sub(invocation.t0)

			//	- invocations may be added to s.invocations
			//		after t timestamp resulting in negative times
			//	- it is a scan for max, so less than zero can
			//		be skipped
			if duration < 0 {
				continue
			}

			// get slow-detector-core for this invocation
			var invoActions = invocation.InvoActions()

			// check if this duration a new max
			if !invoActions.Duration(duration) {
				// this duration is not a new progressive max for this slow-detector
				continue
			}

			// it is a new max, check whether nonReturnPeriod has elapsed
			var tLastNonReturnReport = invocation.Time(time.Time{})
			if !tLastNonReturnReport.IsZero() &&
				t.Sub(tLastNonReturnReport) < s.nonReturnPeriod {
				continue // a previous non-return report is too recent
			}

			// store new nonReturnPeriod start and Report
			invocation.Time(t)
			invoActions.Report(invocation, DidReturnNo, duration)
		}
	}
}

const (
	// how often threads scan for non-return
	slowScanPeriod = time.Second
)

// shared SlowDetectorThread for SlowDefault threads
//   - purpose is fewer threads than slow detectors
var slowDetectorThread SlowDetectorThread
