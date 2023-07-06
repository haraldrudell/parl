/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"testing"

	"github.com/haraldrudell/parl"
)

func ExitingGoroutine(g parl.Go) {
	g.Done(nil)
}

func ReadErrorChannelToEnd(ch <-chan parl.GoError, t *testing.T) {
}

func TestThreadLogger(t *testing.T) {

	// goGroup being logged
	var goGroup parl.GoGroup = NewGoGroup(context.Background())

	// waitgroup for threadLogger end
	var wg = NewThreadLogger(goGroup).Log()

	// launch a quickly terminating goroutine
	//	- this will cause goGroup to terminate
	go ExitingGoroutine(goGroup.Go())

	// read error channel to end
	for {
		e, ok := <-goGroup.Ch()
		if !ok {
			break // error channel closed
		}
		var err = e.Err()
		if err != nil {
			t.Errorf("goGroup err: %s", e)
		}
	}

	t.Log("wg.Wait…")
	wg.Wait()
	t.Log("wg.Wait complete")
}
