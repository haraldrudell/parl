/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package g0

import (
	"context"
	"sync"
	"testing"

	"github.com/haraldrudell/parl"
)

func TestThreadLogger(t *testing.T) {

	var wg = &sync.WaitGroup{}
	defer func() { wg.Wait() }()

	var goGroup parl.GoGroup = NewGoGroup(context.Background())
	goGroup.SetDebug(parl.AggregateThread)

	wg = ThreadLogger(goGroup)

	// launch a quickly terminating goroutine
	//	- this will cause goGroup to terminate
	go func(g0 parl.Go) {
		defer g0.Done(nil)
	}(goGroup.Go())

	for {
		e, ok := <-goGroup.Ch()
		if !ok {
			return // error channel closed
		}
		var err = e.Err()
		if err != nil {
			t.Errorf("goGroup err: %s", e)
		}
	}
}
