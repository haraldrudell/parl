/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

const (
	indicesShouldDetectPanic    = true
	indicesShouldNotDetectPanic = false
	indicesRuntimePrefix        = "runtime."
)

func TestIndices(t *testing.T) {
	newIndicesTest("no panic", noPanic, indicesShouldNotDetectPanic, t).run()
	newIndicesTest("panic(1)", panicOne, indicesShouldDetectPanic, t).run()
	newIndicesTest("nil pointer dereference", panicNilPointer, indicesShouldDetectPanic, t).run()
}

// noPanic returns a deferred stack trace generated without a panic
func noPanic() (stack pruntime.StackSlice) {
	defer func() {
		stack = pruntime.NewStackSlice(0)
	}()
	return
}

// panicOne returns a deferred stack trace generated by panic(1)
func panicOne() (stack pruntime.StackSlice) {
	var one = 1
	defer func() {
		if recoverValue := recover(); recoverValue != one {
			panic(fmt.Errorf("bad recover value: %T “%[1]v” exp: %d",
				recoverValue,
				one,
			))
		}
		stack = pruntime.NewStackSlice(0)
	}()

	panic(one)
}

// panicFunction recovers a panic using [parl.RecoverErr]
//   - panicLine is the exact code line of the panic
//   - err is the error value produced by [parl.RecoverErr]
func panicNilPointer() (stack pruntime.StackSlice) {
	// runtime.errorString “runtime error: invalid memory address or nil pointer dereference”
	//	- runtime.errorString implements error
	//	- only methods are Error() and RuntimeError()
	var message = "runtime error: invalid memory address or nil pointer dereference"
	defer func() {
		var recoverValue = recover()
		var isOk bool
		if err, ok := recoverValue.(error); ok {
			var runtimeError runtime.Error
			if errors.As(err, &runtimeError) {
				isOk = err.Error() == message
			}
		}
		if !isOk {
			panic(fmt.Errorf("bad recover value: %T “%[1]v” exp err message: “%s”",
				recoverValue,
				message,
			))
		}
		stack = pruntime.NewStackSlice(0)
	}()

	// nil pointer dereference panic
	_ = *(*int)(nil)

	return
}

type indicesArgs struct {
	stack pruntime.StackSlice
}

type indicesTest struct {
	name              string
	args              indicesArgs
	wantIsPanic       bool
	wantRecoveryIndex int
	wantPanicIndex    int
	t                 *testing.T
}

func newIndicesTest(name string, stackGenerator func() (stack pruntime.StackSlice), isPanic bool, t *testing.T) (t2 *indicesTest) {
	var tt = indicesTest{
		name:        name,
		args:        indicesArgs{stackGenerator()},
		wantIsPanic: isPanic,
		t:           t,
	}
	t2 = &tt
	if !isPanic {
		return
	}

	// caclulate wantPanicIndex and wantRecoveryIndex
	var hasRecovery bool
	for i := 0; i < len(tt.args.stack); i++ {

		// is this stack frame in the runtime package?
		var isRuntime bool
		var cL *pruntime.CodeLocation = &tt.args.stack[i]
		isRuntime = strings.HasPrefix(cL.FuncName, indicesRuntimePrefix)

		// recovery is the frame before the first runtime-frame
		if !hasRecovery {
			if !isRuntime {
				continue
			}
			hasRecovery = true
			if i > 0 {
				tt.wantRecoveryIndex = i - 1
			}
			continue
		}

		// panic is the frame after the last runtime-frame
		if isRuntime {
			continue
		}
		tt.wantPanicIndex = i
		break
	}

	return
}

func (tt *indicesTest) run() {
	t := tt.t
	t.Run(tt.name, func(t *testing.T) {
		gotIsPanic, gotRecoveryIndex, gotPanicIndex := Indices(tt.args.stack)
		if gotIsPanic != tt.wantIsPanic {
			t.Errorf("Indices() gotIsPanic = %v, want %v", gotIsPanic, tt.wantIsPanic)
		}
		if gotRecoveryIndex != tt.wantRecoveryIndex {
			t.Errorf("Indices() gotRecoveryIndex = %v, want %v", gotRecoveryIndex, tt.wantRecoveryIndex)
		}
		if gotPanicIndex != tt.wantPanicIndex {
			t.Errorf("Indices() gotPanicIndex = %v, want %v", gotPanicIndex, tt.wantPanicIndex)
		}
	})
	if tt.wantIsPanic {
		t.Logf("INPUTSTACK for test: %s stack:%s", tt.name, tt.args.stack)
	}
}