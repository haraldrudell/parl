/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

const (
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
	goGroup  *GoGroup
	log      parl.PrintfFunc
	endCh    chan struct{}
	ErrCh    *parl.NBChan[parl.GoError]
	isCancel atomic.Bool
}

var _ = parl.AggregateThread

// NewThreadLogger wraps a GoGen thread-group in a debug listener
//   - parl.AggregateThread is enabled for the thread-group
//   - ThreadLogger listens for thread-group Cancel
//   - Wait method ensures process does not exit prior to ThreadLogger complete
//   - logFn is an optional logging function, default parl.Log to stderr
//
// Usage:
//
//	main() {
//	  var threadGroup = g0.NewGoGroup(context.Background())
//	  defer threadGroup.Wait()
//	  defer g0.NewThreadLogger(threadGroup).Log().Wait()
//	  defer threadGroup.Cancel()
//	  …
//	 threadGroup.Cancel()
func NewThreadLogger(goGen parl.GoGen, logFn ...func(format string, a ...interface{})) (threadLogger *ThreadLogger) {
	t := ThreadLogger{endCh: make(chan struct{})}

	// obtain logging function
	if len(logFn) > 0 {
		t.log = logFn[0]
	}
	if t.log == nil {
		t.log = parl.Log
	}

	// obtain GoGroup
	var ok bool
	if t.goGroup, ok = goGen.(*GoGroup); !ok {
		panic(perrors.ErrorfPF("type assertion failed, need GoGroup SubGo or SubGroup, received: %T", goGen))
	}
	t.ErrCh = &t.goGroup.ch
	return &t
}

// Log preares the threadgroup for logging on Cancel
func (t *ThreadLogger) Log() (t2 *ThreadLogger) {
	t2 = t

	// if threadGroup has already ended, print that
	var g = t.goGroup
	var log = t.log
	if g.isEnd() {
		log(threadLoggerLabel + ": IsEnd true")
		close(t.endCh)
		return // thread-group already ended
	}
	g.aggregateThreads.Store(true)

	if g.Context().Err() == nil {
		g.goContext.cancelListener = t.cancelListener
		log(threadLoggerLabel + ": listening for Cancel")
		return
	}

	t.launchThread()
	return
}

func (t *ThreadLogger) Wait() {
	<-t.endCh
}

// cancelListener is invoked on every threadGroup.Cancel()
func (t *ThreadLogger) cancelListener() {
	if !t.isCancel.CompareAndSwap(false, true) {
		return // subsequent cancel invocation
	}
	t.log(threadLoggerLabel + ": Cancel detected")
	t.launchThread()
}

// launchThread prepares the waitgroup and lunches the logging thread
func (t *ThreadLogger) launchThread() {
	go t.printThread()
}

// printThread prints goroutines that have yet to exit every second
func (t *ThreadLogger) printThread() {
	var g = t.goGroup
	var log = t.log
	defer close(t.endCh)
	defer parl.Recover(parl.Annotation(), nil, parl.Infallible)
	defer func() { log("%s %s: %s", parl.ShortSpace(), threadLoggerLabel, "thread-group ended") }()

	// ticker for periodic printing
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()

	var endCh <-chan struct{}
	if g.hasErrorChannel.Load() {
		endCh = g.ch.WaitForCloseCh()
	} else {
		endCh = g.endCh.Ch()
	}
	for {

		// multi-line string of all threads
		var threadLines string
		// threads ordered by goEntityID
		var orderedMap = g.ThreadsInternal()
		// printable string representation of all threads
		var ts = make([]string, orderedMap.Length())
		for i, goEntityId := range orderedMap.List() {
			var threadData, _ = orderedMap.Get(goEntityId)
			var _ parl.ThreadData
			ts[i] = threadData.(*ThreadData).LabeledString() + " G" + goEntityId.String()
		}
		threadLines = strings.Join(ts, "\n")

		// header line
		//	- 230622 16:51:28-07 ThreadLogger: GoGen: goGroup#1_threads:316(325)_
		//		New:main.main()-graffick.go:111 threads: 317
		log("%s %s: GoGen: %s threads: %d\n%s",
			parl.ShortSpace(),   // 230622 16:51:26-07
			threadLoggerLabel,   // ThreadLogger
			g,                   // GoGen: goGroup#1…
			orderedMap.Length(), // threads: 317
			threadLines,         // one line for each thread
		)

		// exit if thread-group done
		if g.isEnd() {
			return
		}

		// blocks here
		select {
		case <-endCh: // thread-group ended
		case <-ticker.C: // timer trig
		}
	}
}
