/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

func TestRecoverDA(t *testing.T) {
	// expFormat recreates panic formatting
	var expFormat = "panic detected in %s: “%s” at %s"
	// the panic message for the type of panic carried out
	//	- “runtime error: invalid memory address or nil pointer dereference”
	var expPanicMessage = func() (message string) {
		defer func() {
			message = recover().(error).Error()
		}()

		_ = *(*int)(nil)
		return
	}()

	// expPanicMessage: "runtime error: invalid memory address or nil pointer dereference"
	t.Logf("expPanicMessage: %q", expPanicMessage)

	// deferCL is the line of function recoverDaPanic where deferLocation is assigned
	var deferCL *pruntime.CodeLocation
	// panicCL is the line in panickingFunction where panicLine is assigned
	var panicCL *pruntime.CodeLocation
	var err error
	var message string

	deferCL, panicCL, err = recoverDaPanic()

	// should be error
	if err == nil {
		t.Fatalf("missing error")
	}
	// error-chain: *fmt.wrapError *errorglue.errorStack runtime.errorString
	t.Logf("error-chain: %s", errorglue.DumpChain(err))

	// defer location: parl.recoverDaPanic()-recover3_test.go:16
	t.Logf("defer location: %s", deferCL.Short())
	// panic location: parl.panickingFunction()-recover3_test.go:24
	t.Logf("panic location: %s", panicCL.Short())

	var expMessage = fmt.Sprintf(expFormat,
		deferCL.PackFunc(), // parl.recoverDaPanic
		expPanicMessage,    // runtime error…
		panicCL.Short(),
	)
	message = perrors.Short(err)
	if message != expMessage {
		t.Errorf("FAIL bad message:\n%q exp\n%q",
			message,
			expMessage,
		)
	}
}

// recovers a panic in a called function
//   - deferLocation is the function where ercovery takes place
//   - panicLocation is the called function where the panic occurred
//   - err is the resultfrom [Recover]
func recoverDaPanic() (deferLocation, panicLocation *pruntime.CodeLocation, err error) {
	deferLocation = pruntime.NewCodeLocation(0)
	defer Recover(func() DA { return A() }, &err, NoOnError)

	panickingFunction(&panicLocation)
	return
}

// provides a code location on the same line as a panic is caused
func panickingFunction(panicLine **pruntime.CodeLocation) {
	if *panicLine = pruntime.NewCodeLocation(0); *(*int)(nil) != 0 {
		_ = 1
	}
}
