/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
	"time"

	"github.com/haraldrudell/parl/ptime"
	"golang.org/x/exp/slices"
)

func TestDebouncer(t *testing.T) {
	var value1 = 3
	var value2 = 5
	var debouncePeriod = time.Millisecond
	var noDebounce time.Duration
	var noMaxDelay time.Duration
	var maxDelay1ms = time.Millisecond
	var expValues = []int{value1, value2}
	var expValues1 = []int{value1}

	var actValues []int
	var t0, t1 time.Time

	// create debouncer immediately receiving 2 values
	var inputCh = make(chan int, 2)
	inputCh <- value1
	inputCh <- value2
	var receiver AwaitableSlice[[]int]
	t0 = time.Now()
	var debouncer = NewDebouncer(
		debouncePeriod,
		noMaxDelay,
		inputCh,
		receiver.Send,
		newPanicOnError(),
	)

	// actValues should receive a slice of two values
	//	- received because of debounce timer 1 ms
	<-receiver.DataWaitCh()
	actValues, _ = receiver.Get()
	t1 = time.Now()
	t.Logf("elapsed debounce: %s", ptime.Duration(t1.Sub(t0)))
	if !slices.Equal(actValues, expValues) {
		t.Errorf("bad receive: %v exp %v", actValues, expValues)
	}

	// Shutdown should not panic or hang
	t.Log("debouncer.Shutdown…")
	debouncer.Shutdown()
	t.Log("debouncer.Shutdown complete")

	// create a debouncer receiving no values
	t0 = time.Now()
	debouncer = NewDebouncer(
		debouncePeriod,
		noMaxDelay,
		inputCh,
		receiver.Send,
		newPanicOnError(),
	)

	// close of input channel should terminate the debouncer
	close(inputCh)
	t.Log("debouncer.Wait…")
	debouncer.Wait()
	t1 = time.Now()
	t.Logf("debouncer.Wait complete: %s", ptime.Duration(t1.Sub(t0)))

	// create a debouncer with maxDelay
	inputCh = make(chan int, 1)
	inputCh <- value1
	t0 = time.Now()
	debouncer = NewDebouncer(
		noDebounce,
		maxDelay1ms,
		inputCh,
		receiver.Send,
		newPanicOnError(),
	)

	// maxDelay should release one value
	<-receiver.DataWaitCh()
	actValues, _ = receiver.Get()
	t1 = time.Now()
	t.Logf("elapsed maxDelay: %s", ptime.Duration(t1.Sub(t0)))
	if !slices.Equal(actValues, expValues1) {
		t.Errorf("bad receive: %v exp %v", actValues, expValues1)
	}

	// shutdown should not panic or hang
	debouncer.Shutdown()

	//t.Fail()
}

type panicOnError struct{}

func newPanicOnError() (errorSink ErrorSink) { return NewErrorSinkEndable(&panicOnError{}) }

// debouncerPanicErrFn is a debouncer errFN that panics
//   - debuncer does not have any errors
func (p *panicOnError) AddError(err error) {
	panic(err)
}
