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
	const (
		value1, value2 = 3, 5
		debouncePeriod = time.Millisecond
		maxDelay1ms    = time.Millisecond
		valueCount     = 2
	)
	var (
		expValues  = []int{value1, value2}
		expValues1 = []int{value1}
	)

	var (
		inputCh    chan int
		noDebounce time.Duration
		noMaxDelay time.Duration
		actValues  []int
		t0, t1     time.Time
		receiver   AwaitableSlice[[]int]
		hasValue   bool
	)

	// create debouncer immediately receiving 2 values
	inputCh = make(chan int, valueCount)
	inputCh <- value1
	inputCh <- value2
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
	actValues, hasValue = receiver.Get()
	t1 = time.Now()

	// elapsed debounce: 1.284ms
	t.Logf("elapsed debounce: %s debouncePeriod: %s",
		ptime.Duration(t1.Sub(t0)),
		debouncePeriod,
	)
	if !hasValue {
		t.Errorf("on dataWait hasValue false")
	}
	if !slices.Equal(actValues, expValues) {
		t.Errorf("FAIL bad receive: %v exp %v", actValues, expValues)
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

	// debouncer.Wait complete: 21µs
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
	actValues, hasValue = receiver.Get()
	t1 = time.Now()

	// elapsed maxDelay: 1.216ms maxDelay: 1ms
	t.Logf("elapsed maxDelay: %s maxDelay: %s",
		ptime.Duration(t1.Sub(t0)),
		maxDelay1ms,
	)
	if !hasValue {
		t.Errorf("on dataWait hasValue false")
	}
	if !slices.Equal(actValues, expValues1) {
		t.Errorf("FAIL bad receive: %v exp %v", actValues, expValues1)
	}

	// shutdown should not panic or hang
	debouncer.Shutdown()
}

// panicOnError is an errorSink that panics
type panicOnError struct{}

// newPanicOnError returns an errorSink that panics
func newPanicOnError() (errorSink ErrorSink) { return NewErrorSinkEndable(&panicOnError{}) }

// debouncerPanicErrFn is a debouncer errFN that panics
//   - debuncer does not have any errors
func (p *panicOnError) AddError(err error) {
	panic(err)
}
