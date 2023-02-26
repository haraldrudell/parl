/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"sync"
	"testing"
)

func TestAwaitableCalculation(t *testing.T) {
	value := 1

	var calculation *Future[int]
	var result int
	var isValid bool

	calculation = NewFuture[int]()

	var wgStart sync.WaitGroup
	wgStart.Add(1)
	var wgEnd sync.WaitGroup
	wgEnd.Add(1)
	go func() {
		defer wgEnd.Done()
		wgStart.Done()
		result, isValid = calculation.Result()
	}()
	wgStart.Wait()

	if calculation.IsCompleted() {
		t.Error("calculation.IsCompleted true")
	}

	calculation.End(&value, nil)
	if !calculation.IsCompleted() {
		t.Error("calculation.IsCompleted false")
	}

	wgEnd.Wait()
	if !isValid {
		t.Error("isValid false")
	}
	if result != value {
		t.Errorf("result bad: %d exp %d", result, value)
	}
}
