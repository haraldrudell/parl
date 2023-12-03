/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

func TestRecoverDA(t *testing.T) {
	var expFormat = "panic detected in %s: “%s” at %s"
	var expPanicMessage = func() (message string) {
		defer func() {
			message = recover().(error).Error()
		}()

		_ = *(*int)(nil)
		return
	}()

	deferCL, panicCL, err := recoverDaPanic()

	// should be error
	if err == nil {
		t.Error("missing error")
		t.FailNow()
	}

	// defer location: parl.recoverDaPanic()-recover3_test.go:16
	t.Logf("defer location: %s", deferCL.Short())

	// panic location: parl.panickingFunction()-recover3_test.go:24
	t.Logf("panic location: %s", panicCL.Short())

	var expMessage = fmt.Sprintf(expFormat,
		deferCL.PackFunc(),
		expPanicMessage,
		panicCL.Short(),
	)
	var message = perrors.Short(err)
	if message != expMessage {
		t.Errorf("bad message:\n%q exp\n%q",
			message,
			expMessage,
		)
	}
}

func recoverDaPanic() (deferLocation, panicLocation *pruntime.CodeLocation, err error) {
	deferLocation = pruntime.NewCodeLocation(0)
	defer Recover(func() DA { return A() }, &err, NoOnError)

	panickingFunction(&panicLocation)
	return
}

func panickingFunction(panicLine **pruntime.CodeLocation) {
	if *panicLine = pruntime.NewCodeLocation(0); *(*int)(nil) != 0 {
		_ = 1
	}
}
