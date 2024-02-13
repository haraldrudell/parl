/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"slices"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

func TestAddNotifier(t *testing.T) {
	var countersCount = 4
	var expCounters = []int{1, 0, 1, 1}

	// check expCounters
	if len(expCounters) != countersCount {
		panic(perrors.NewPF("bad expcounters length"))
	}

	var ctx context.Context
	var actCounters = make([]int, countersCount)

	// get packFunc
	var packFunc = notifierInvokeCancel(ctx)
	// packFunc: parl.notifierInvokeCancel
	t.Logf("packFunc: %s", packFunc)

	// create counters
	var counters = make([]*notifierCounter, countersCount)
	for i := range counters {
		counters[i] = newNotifierCounter(packFunc, t)
	}

	// create contexts ctx…ctx4
	ctx = context.Background()
	// counters[0] is notified of all context cancels
	var ctx0 = AddNotifier(ctx, counters[0].notifierFunc)
	// counters[1] is notified of context cancels in
	// child contexts without a notifier1
	//	- ie. no cancelations
	var ctx1 = AddNotifier1(ctx0, counters[1].notifierFunc)
	// counters[2] is notified of context cancels in ctx2…
	var ctx2 = AddNotifier1(ctx1, counters[2].notifierFunc)
	// counters[3] is notified of all context cancels
	var ctx3 = AddNotifier(ctx2, counters[3].notifierFunc)
	// ctx4 is CancelContext
	var ctx4 = NewCancelContext(ctx3)

	// cancel a context
	notifierInvokeCancel(ctx4)

	// counter invocations should match expCounters
	for i, c := range counters {
		actCounters[i] = c.count
	}
	if !slices.Equal(actCounters, expCounters) {
		t.Errorf("counters bad:\n%v exp:\n%v",
			actCounters,
			expCounters,
		)
	}
}

// notifierCounter is a fixture counting invocations
type notifierCounter struct {
	count    int
	packFunc string
	t        *testing.T
}

// newNotifierCounter returns a notifier that counts its invocations
func newNotifierCounter(packFunc string, t *testing.T) (c *notifierCounter) {
	return &notifierCounter{
		packFunc: packFunc,
		t:        t,
	}
}

// notifierFunc is a notifierFunc function for child or all contexts
func (c *notifierCounter) notifierFunc(stack Stack) {
	t := c.t

	// count invocation
	c.count++

	// check stack trace
	var frames = stack.Frames()
	if len(frames) < 2 {
		panic(perrors.ErrorfPF("bad stack slice: %s"))
	}
	var tracePackFunc = frames[1].Loc().PackFunc()
	if tracePackFunc == c.packFunc {
		return // stack trace OK return
	}
	t.Logf("TRACE: %s", stack)
	panic(perrors.New("Bad stack slice"))
}

// notifierInvokeCancel retrieves packFunc and invokes InvokeCancel
func notifierInvokeCancel(ctx context.Context) (packFunc string) {
	if ctx != nil {
		// if ctx present, cancel it
		InvokeCancel(ctx)
	} else {
		// if ctx not present, return packFunc for cancellation stack trace
		packFunc = pruntime.NewCodeLocation(0).PackFunc()
	}
	return
}
