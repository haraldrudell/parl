/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pslices

import (
	"testing"
)

func TestRangeCh(t *testing.T) {
	var value1 = 1
	var value2 = 2

	var threadSafeSlice = NewThreadSafeSlice[int]()
	threadSafeSlice.Append(value1)
	threadSafeSlice.Append(value2)
	var ch <-chan int
	var actual int
	var ok bool

	var rangeCh = NewRangeCh(threadSafeSlice)
	ch = rangeCh.Ch()

	// read first value
	if actual, ok = <-ch; !ok {
		t.Error("1: ok: false")
	}
	if actual != value1 {
		t.Errorf("Bad actual 1: %d exp %d", actual, value1)
	}

	rangeCh.Close()

	if actual, ok = <-ch; ok {
		t.Error("2: ok: true")
	}
	_ = actual
}
