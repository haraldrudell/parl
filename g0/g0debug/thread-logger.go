/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0debug

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/g0"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps"
	"github.com/haraldrudell/parl/pslices"
)

const (
	// “ThreadLogger” preceeds all thread-logger output
	threadLoggerLabel = "ThreadLogger"
)

// ThreadLogger waits for a GoGroup, SubGo or SubGroup to terminate while printing
// information on threads that have yet to exit every second.
//   - Because the GoGroup owner needs to continue consuming the GoGroup’s error channel,
//     ThreadLogger has built-in threading
//   - the returned sync.WaitGroup pointer should be used to ensure main does
//     not exit prematurely. The WaitGroup ends when the GoGroup ends and ThreadLogger
//     ceases output
type ThreadLogger struct {
	endCh    chan struct{}
	isCancel atomic.Bool
	// the log output is printed to
	log parl.PrintfFunc

	// goGroup is thread-group being monitored
	goGroup *g0.GoGroup
	// isEnd is the thread-group’s private method returning end status
	isEnd func() bool
	// isAggregateThreads is the thread-group’s private flag for whether to collect thread data
	isAggregateThreads *atomic.Bool
	// setCancelListener is the thread-group’s private method for installing a listener to context cancel
	setCancelListener func(f func())
	// gEndCh is the thread-group’s private end channel for SubGo that does not have an error stream
	gEndCh <-chan struct{}
}

var _ = parl.AggregateThread

// NewThreadLogger provides debug logging for a thread-group
//   - GoGen: an thread-group object managing threads implemented by [g0.GoGroup]:
//   - — [parl.GoGroup]
//   - — [parl.Subgo]
//   - — [parl.SubGroup]
//   - logFn: an optional logging function, default [parl.Log] to stderr
//   - the new function does not take any action or prepare logging,
//     actions begin by invoking the Log method
//   - [ThreadLogger.Log] begins monitoring and should be invoked prior to launching any threads
//   - — ThreadLogger enables [parl.AggregateThread] for the thread-group.
//     This causes the thread-group to collect thread information for debug purposes
//   - — ThreadLogger then listens for thread-group Cancel allowing logging to start automatically
//   - [ThreadLogger.Wait] awaits the end of the thread-group and the thread logger
//   - — Wait should not be invoked prior to ensuring that the thread-group is shutting down
//   - — a thread-group is typically shut down via its Cancel method,
//     but a thread-group also shuts down on its last thread exiting,
//     a fatal thread error or for other reasons
//   - — invoking the Wait method ensures the process does not exit prior to ThreadLogger complete
//   - ThreadLogger uses a thread for logging that exits upon the thread-group ending
//
// Usage:
//
//	main() {
//	  var threadGroup = g0.NewGoGroup(context.Background())
//	  defer threadGroup.Wait()
//	  defer g0debug.NewThreadLogger(threadGroup).Log().Wait()
//	  defer threadGroup.Cancel()
//	  …
//	 threadGroup.Cancel()
func NewThreadLogger(goGen parl.GoGen, logFn ...parl.PrintfFunc) (threadLogger *ThreadLogger) {
	t := ThreadLogger{endCh: make(chan struct{})}

	// obtain logging function
	if len(logFn) > 0 {
		t.log = logFn[0]
	}
	if t.log == nil {
		t.log = parl.Log
	}

	// obtain implementing GoGroup
	var ok bool
	if t.goGroup, ok = goGen.(*g0.GoGroup); !ok {
		panic(perrors.ErrorfPF("type assertion failed, need GoGroup SubGo or SubGroup, received: %T", goGen))
	}

	// retrieve internals from the goGroup
	t.isEnd, t.isAggregateThreads, t.setCancelListener, t.gEndCh, _ /*goError*/ = t.goGroup.Internals()

	return &t
}

// Log prepares the threadgroup for logging on Cancel
//   - Log activates a Cancel listener on the thread-group
//     allowing thread-logging to start automatically
//   - supports functional chaining
func (t *ThreadLogger) Log() (t2 *ThreadLogger) {
	t2 = t

	// if threadGroup has already ended, print that
	var g = t.goGroup
	var log = t.log
	if t.isEnd() {
		log(threadLoggerLabel + ": IsEnd true")
		close(t.endCh)
		return // thread-group already ended
	}
	t.isAggregateThreads.Store(true)

	if g.Context().Err() == nil {
		t.setCancelListener(t.cancelListener)
		log(threadLoggerLabel + ": listening for Cancel")
		return
	}

	t.launchThread()
	return
}

// Wait awaits the monitored thread-group end which ends the thread-logger, too
func (t *ThreadLogger) Wait() { <-t.endCh }

// cancelListener is invoked on every threadGroup.Cancel()
func (t *ThreadLogger) cancelListener() {
	if !t.isCancel.CompareAndSwap(false, true) {
		return // subsequent cancel invocation
	}
	t.log(threadLoggerLabel + ": Cancel detected")
	t.launchThread()
}

// launchThread prepares the waitgroup and lunches the logging thread
func (t *ThreadLogger) launchThread() { go t.printThread() }

// printThread prints goroutines that have yet to exit every second
func (t *ThreadLogger) printThread() {
	var g = t.goGroup
	var log = t.log
	defer close(t.endCh)
	defer parl.Recover(func() parl.DA { return parl.A() }, nil, parl.Infallible)
	defer func() { log("%s %s: %s", parl.ShortSpace(), threadLoggerLabel, "thread-group ended") }()

	// ticker for periodic printing
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()

	for {

		// multi-line string of all threads
		var threadLines string

		// unsorted map:
		//	- key: internal parl.GoEntityID
		//	- value: *ThreadData, has no GoEntityID
		//	- keys must be retrieved for order
		//	- values must be retrieved for printing
		var m = g.ThreadsInternal() // unordered list parl.ThreadData
		// get implementation that has Range method
		var rwm = m.(*pmaps.RWMap[parl.GoEntityID, *g0.ThreadData])
		// assemble sorted list of keys
		var goEntityOrder = make([]parl.GoEntityID, m.Length())[:0]
		rwm.Range(func(key parl.GoEntityID, value *g0.ThreadData) (keepGoing bool) {
			goEntityOrder = pslices.InsertOrdered(goEntityOrder, key)
			return true
		})

		// printable string representation of all threads
		var ts = make([]string, len(goEntityOrder))
		for i, goEntityId := range goEntityOrder {
			var threadData, _ = m.Get(goEntityId)
			ts[i] = threadData.LabeledString() + " G" + goEntityId.String()
		}
		threadLines = strings.Join(ts, "\n")

		// header line
		//	- 230622 16:51:28-07 ThreadLogger: GoGen: goGroup#1_threads:316(325)_
		//		New:main.main()-graffick.go:111 threads: 317
		log("%s %s: GoGen: %s threads: %d\n%s",
			parl.ShortSpace(),  // 230622 16:51:26-07
			threadLoggerLabel,  // ThreadLogger
			g,                  // GoGen: goGroup#1…
			len(goEntityOrder), // threads: 317
			threadLines,        // one line for each thread
		)

		// exit if thread-group done
		if t.isEnd() {
			return
		}

		// blocks here
		select {
		case <-t.gEndCh: // thread-group ended
		case <-ticker.C: // timer trig
		}
	}
}
