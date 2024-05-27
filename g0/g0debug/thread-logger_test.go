/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0debug

import (
	"context"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/g0"
)

func TestThreadLogger(t *testing.T) {

	var (
		// goGroup being logged
		goGroup  parl.GoGroup = g0.NewGoGroup(context.Background())
		goErrors parl.IterableSource[parl.GoError]
		hasValue bool
		goError  parl.GoError
		endCh    parl.AwaitableCh
	)

	// Log Wait
	var threadLogger *ThreadLogger = NewThreadLogger(goGroup)

	// arm logging for the thread-group
	threadLogger.Log()

	// launch a quickly terminating goroutine
	//	- this will cause goGroup to terminate
	//	- logging will then start
	go exitingGoroutine(goGroup.Go())

	// read error channel to end
	//	- this will cause the thread-group to end
	goErrors = goGroup.GoError()
	endCh = goErrors.EmptyCh(parl.CloseAwaiter)
	for {
		select {
		case <-goErrors.DataWaitCh():
			if goError, hasValue = goErrors.Get(); !hasValue {
				continue
			} else if goError == nil {
				continue
			} else if goError.Err() != nil {
				t.Errorf("goGroup err: %s", goError)
			}
		case <-endCh:
		}
		break
	}

	// await thread logger exit
	t.Log("wg.Wait…")
	threadLogger.Wait()
	t.Log("wg.Wait complete")
}

// exitingGoroutine is a goroutine immediately exiting successfully
func exitingGoroutine(g parl.Go) { g.Done(nil) }
