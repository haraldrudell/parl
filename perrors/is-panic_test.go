/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

func TestIsPanic(t *testing.T) {
	message := "message"
	err := New(message)

	var isPanic bool
	var stack pruntime.Stack
	var recoveryIndex int
	var panicIndex int
	errPanic := func() (err error) {
		defer func() {
			recover()
			err = New("panic")
		}()
		panic(1)
	}()

	isPanic, _, _, _ = IsPanic(err)
	if isPanic {
		t.Error("non-panic error has isPanic true")
	}

	isPanic, stack, recoveryIndex, panicIndex = IsPanic(errPanic)
	var frames = stack.Frames()
	t.Logf("isPanic: %t stack: %d recoveryIndex: %d panicIndex: %d",
		isPanic, len(frames), recoveryIndex, panicIndex)
	t.Logf("\nstack:%s", stack)
	if !isPanic {
		t.Error("panic error has isPanic false")
	}
	if len(frames) == 0 {
		t.Error("stack length zero exp non-zero")
	}
	// recoveryIndex cn be 0
	if panicIndex == 0 {
		t.Error("panicIndex zero exp non-zero")
	}
}
