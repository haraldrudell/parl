/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
	"time"
)

func TestNewWinOrWaiterCore(t *testing.T) {
	var winOrWaiterStrategy = WinOrWaiterAnyValue
	var shortTime = time.Millisecond

	var calculatorInvoked = NewMutexWait()
	var completeCalculation = NewMutexWait()
	var calculator = func() (err error) {
		calculatorInvoked.Unlock()
		completeCalculation.Wait()
		return
	}
	var calculationReference *Future[time.Time]
	var result time.Time
	var isValid bool

	var winOrWaiter *WinOrWaiterCore = NewWinOrWaiterCore(winOrWaiterStrategy, calculator)

	if winOrWaiter.strategy != winOrWaiterStrategy {
		t.Errorf("bad startegy: %d exp %d", winOrWaiter.strategy, winOrWaiterStrategy)
	}
	if winOrWaiter.isCalculationPut.Load() {
		t.Error("winOrWaiter.isCalculationPut true")
	}
	if winOrWaiter.calculation.Load() != nil {
		t.Error("winOrWaiter.calculation not nil")
	}
	if winOrWaiter.winnerPicker.Load() {
		t.Error("winOrWaiter.winnerPicker true")
	}

	// first thread gets stuck in calculator func
	var calculatorThread = NewMutexWait()
	go func() {
		defer calculatorThread.Unlock()

		winOrWaiter.WinOrWait()
	}()
	t.Log("waiting for calculator invocation…")
	calculatorInvoked.Wait()

	if !winOrWaiter.isCalculationPut.Load() {
		t.Error("2 winOrWaiter.isCalculationPut false")
	}
	if winOrWaiter.calculation.Load() == nil {
		t.Error("2 winOrWaiter.calculation nil")
	}
	if !winOrWaiter.winnerPicker.Load() {
		t.Error("2 winOrWaiter.winnerPicker false")
	}

	// second thread gets stuck in at ww.calculator.RWMutex
	var loserThread = NewMutexWait()
	var loserThreadUp = NewMutexWait()
	go func() {
		defer loserThread.Unlock()

		loserThreadUp.Unlock()
		winOrWaiter.WinOrWait()
	}()
	t.Log("waiting for loser thread…")
	loserThreadUp.Wait()

	time.Sleep(shortTime)

	if calculatorThread.IsUnlocked() {
		t.Error("calculatorThread exited too soon")
	}
	if loserThread.IsUnlocked() {
		t.Error("loserThread exited too soon")
	}

	completeCalculation.Unlock()

	calculatorThread.Wait()
	loserThread.Wait()

	calculationReference = winOrWaiter.calculation.Load()
	if calculationReference == nil {
		t.Error("3 winOrWaiter.calculation.Get nil")
		t.FailNow()
	}
	if !calculationReference.IsCompleted() {
		t.Error("3 calculationReference.IsCompleted false")
	}
	result, isValid = calculationReference.Result()
	if !isValid {
		t.Error("3 isValid false")
	}
	if result.IsZero() {
		t.Error("3 result IsZero")
	}
	if winOrWaiter.winnerPicker.Load() {
		t.Error("3 winOrWaiter.winnerPicker true")
	}
}
