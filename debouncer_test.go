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

	var inputCh = make(chan int, 2)
	inputCh <- value1
	inputCh <- value2
	var receiver NBChan[[]int]
	var errFn = func(err error) {
		panic(err)
	}
	var v []int
	var expV = []int{value1, value2}

	var t0 = time.Now()
	var debouncer = NewDebouncer(
		debouncePeriod,
		inputCh,
		receiver.Send,
		errFn,
	).Go()

	v = <-receiver.Ch()
	var t1 = time.Now()
	t.Logf("elapsed: %s", ptime.Duration(t1.Sub(t0)))
	if !slices.Equal(v, expV) {
		t.Errorf("bad receive: %v exp %v", v, expV)
	}
	t.Log("debouncer.Shutdown…")
	debouncer.Shutdown()
	t.Log("debouncer.Shutdown complete")

	debouncer = NewDebouncer(
		debouncePeriod,
		inputCh,
		receiver.Send,
		errFn,
	).Go()
	close(inputCh)
	t.Log("debouncer.Wait…")
	debouncer.Wait()
	t.Log("debouncer.Wait complete")
}
