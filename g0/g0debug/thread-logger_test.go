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
		goGroup parl.GoGroup = g0.NewGoGroup(context.Background())
		hadExit bool
	)

	// Log() Wait()
	var threadLogger *ThreadLogger = NewThreadLogger(goGroup)

	// arm logging for the thread-group
	//	- logging starts on thread-group Cancel
	threadLogger.Log()

	// launch a quickly terminating goroutine
	//	- this will cause goGroup to terminate
	//	- logging will then start
	go exitingGoroutine(goGroup.Go())

	// read error channel to end
	//	- this will cause the thread-group to end
	for goError := range goGroup.GoError().Seq {

		// there should be one GeExit with nil error
		if goError != nil &&
			!hadExit &&
			goError.ErrContext() == parl.GeExit &&
			goError.Err() == nil {
			hadExit = true
			continue
		}

		t.Errorf("goGroup err: %s", GoErrorDump(goError))
	}
	if !hadExit {
		t.Error("no thread exit")
	}

	// await thread logger exit
	t.Log("wg.Wait…")
	threadLogger.Wait()
	t.Log("wg.Wait complete")
}

// exitingGoroutine is a goroutine immediately exiting successfully
func exitingGoroutine(g parl.Go) { g.Done(nil) }
