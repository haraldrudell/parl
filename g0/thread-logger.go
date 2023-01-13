/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"fmt"
	"strings"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

type goGroup interface {
	isEnd() (isEnd bool)
	listThreads() (threads []*ThreadData)
	Wait()
	fmt.Stringer
}

// ThreadLogger waits for a GoGroup, SubGo or SubGroup while printing
// information on threads that have yet to exit.
//   - ThreadLogger is non-blocking and will invoke Done on wg
//
// Usage:
//
//	main() {
//	  var debugWait sync.WaitGroup
//	  defer debugWait.Wait()
//	 …
//	 debugWait.Add(1)
//	 ThreadLogger(goGroup, &debugWait)
func ThreadLogger(goGen parl.GoGen, wg parl.SyncDone, logFn ...func(format string, a ...interface{})) {

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

	if g0.isEnd() {
		log("ThreadLogger: IsEnd true")
		wg.Done()
		return // thread-group already ended
	}

	go printThread(wg, log, g0)
}

func printThread(wg parl.SyncDone, log parl.PrintfFunc, g0 goGroup) {
	wg.Done()
	defer parl.Recover(parl.Annotation(), nil, parl.Infallible)

	// channel indicating Wait complete
	waitCh := make(chan struct{})
	go threadLoggerThread(g0, waitCh)

	// ticker for periodic printing
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	defer func() {
		log("%s %s", parl.ShortSpace(), "thread-group ended")
	}()
	for {
		threads := g0.listThreads()
		ts := make([]string, len(threads))
		for i, t := range threads {
			ts[i] = parl.Sprintf("threadID:%s go:%s", t.threadID, t.createLocation.Short())
		}
		s := strings.Join(ts, "\n")
		log("%s %s\n%s", parl.ShortSpace(), g0.String(), s)
		select {
		case <-waitCh:
			return
		case <-ticker.C:
		}
	}
}

func threadLoggerThread(g0 goGroup, waitCh chan struct{}) {
	defer parl.Recover(parl.Annotation(), nil, parl.Infallible)

	g0.Wait()
	close(waitCh)
}
