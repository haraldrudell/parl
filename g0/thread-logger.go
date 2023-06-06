/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pmaps"
	"github.com/haraldrudell/parl/pruntime"
)

type goGroup interface {
	isEnd() (isEnd bool)
	ThreadsInternal() (orderedMap pmaps.KeyOrderedMap[GoEntityID, parl.ThreadData])
	Context() (ctx context.Context)
	fmt.Stringer
}

var c0 pruntime.CachedLocation

// ThreadLogger waits for a GoGroup, SubGo or SubGroup to terminate while printing
// information on threads that have yet to exit every second.
//   - invoke ThreadLogger after goGroup.Cancel and ThreadLogger will list information
//     on goroutines that has yet to exit
//   - goGen is the thread group and is a parl.GoGroup SubGo or SubGroup
//   - — goGen should have SetDebug(parl.AggregateThread) causing it to aggregate
//     information on live threads
//   - logFn is an optional logging function, default parl.Log to stderr
//   - ThreadLogger returns pointer to sync.WaitGroup so it can be waited upon
//   - —
//   - Because the GoGroup owner needs to continue consuming the GoGroup’s error channel,
//     ThreadLogger has built-in threading
//   - the returned sync.WaitGroup pointer should be used to ensure main does
//     not exit prematurely. The WaitGroup ends when the GoGroup ends and ThreadLogger
//     ceases output
//
// Usage:
//
//	main() {
//	  var wg = &sync.WaitGroup{}
//	  defer func() { wg.Wait() }()
//	  var goGroup = g0.NewGoGroup(context.Background())
//	  goGroup.SetDebug(parl.AggregateThread)
//	 …
//	 goGroup.Cancel()
//	 wg = ThreadLogger(goGroup)
func ThreadLogger(goGen parl.GoGen, logFn ...func(format string, a ...interface{})) (
	wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}

	// obtain logging function
	var log parl.PrintfFunc
	if len(logFn) > 0 {
		log = logFn[0]
	}
	if log == nil {
		log = parl.Log
	}

	// obtain GoGroup
	var g0 goGroup
	var ok bool
	if g0, ok = goGen.(goGroup); !ok {
		panic(perrors.ErrorfPF("type assertion failed, need GoGroup SubGo or SubGroup, received: %T", goGen))
	}

	// wait for g0 to end with logging to log
	if g0.isEnd() {
		log("%s: IsEnd true", c0.FuncIdentifier())
		return // thread-group already ended
	}
	wg.Add(1)
	go printThread(wg, log, c0.FuncIdentifier(), g0)
	return
}

// printThread prints goroutines that have yet to exit every second
func printThread(wg parl.SyncDone, log parl.PrintfFunc, label string, g0 goGroup) {
	defer wg.Done()
	defer parl.Recover(parl.Annotation(), nil, parl.Infallible)
	defer func() { log("%s %s: %s", parl.ShortSpace(), label, "thread-group ended") }()

	// ticker for periodic printing
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		var orderedMap = g0.ThreadsInternal()
		ts := make([]string, orderedMap.Length())
		for i, goEntityId := range orderedMap.List() {
			var threadData, _ = orderedMap.Get(goEntityId)
			var _ parl.ThreadData
			ts[i] = threadData.(*ThreadData).LabeledString() + " G" + goEntityId.String()
		}
		threadLines := strings.Join(ts, "\n")
		log("%s %s: GoGen: %s threads: %d\n%s",
			parl.ShortSpace(),
			label,
			g0,
			len(threadLines), threadLines,
		)

		// blocks here
		select {
		case <-g0.Context().Done():
			return
		case <-ticker.C:
		}
	}
}
