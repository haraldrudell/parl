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
	var (
		value = 1
	)

	var (
		result  int
		isValid bool
		wgStart sync.WaitGroup
		wgEnd   sync.WaitGroup
	)

	checkAllocation()

	// Ch() End() IsCompleted() Result() TResult()
	var calculation Future[int]

	// prior to End IsCompleted should be false
	// create and await start of calculation thread
	wgStart.Add(1)
	wgEnd.Add(1)
	go awaitCalculationThread(&calculation, &result, &isValid, &wgStart, &wgEnd)
	wgStart.Wait()
	if calculation.IsCompleted() {
		t.Error("calculation.IsCompleted true")
	}
	if isValid {
		t.Error("isValid true")
	}

	// after End, IsCompleted should be true
	calculation.End(&value, nil, nil)
	if !calculation.IsCompleted() {
		t.Error("calculation.IsCompleted false")
	}

	// Result should provide the correct values
	// await the thread to provide value
	wgEnd.Wait()
	if !isValid {
		t.Error("isValid false")
	}
	if result != value {
		t.Errorf("result bad: %d exp %d", result, value)
	}
}

// awaitCalculationThread flags start end and carries out the calculation
func awaitCalculationThread(calculation *Future[int], result *int, isValid *bool, wgStart, wgEnd *sync.WaitGroup) {
	defer wgEnd.Done()

	wgStart.Done()
	var resultp *int
	resultp, *isValid = calculation.Result()
	*result = *resultp
}

// checkAllocation can be used to inspect
// if stack-allocation is possible
//   - not possible 220524
func checkAllocation() {
	var c Future[int]
	var i = 1

	c.End(&i, NoIsPanic, NoErrp)
}
