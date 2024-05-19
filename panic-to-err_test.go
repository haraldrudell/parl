/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"fmt"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

func TestPanicToErr(t *testing.T) {
	// "runtime error: invalid memory address or nil pointer dereference"
	var panicMessage = func() (message string) {
		defer func() { message = recover().(error).Error() }()
		var intp *int
		if false {
			var i int
			intp = &i
		}
		_ = *intp // causes nil pointer dereference panic
		return
	}()
	// “runtime error: invalid memory address or nil pointer dereference”
	var expMessage = fmt.Sprintf("“%s”", panicMessage)

	var panicLine *pruntime.CodeLocation
	var err error
	var stack pruntime.Stack
	var errorShort string

	// get [parl.RecoverErr] values from recovering a panic
	panicLine, err = panicFunction()

	// there should be an error
	if err == nil {
		t.Error("expected error missing")
		t.FailNow()
	}
	stack = errorglue.GetStackTrace(err)
	errorShort = perrors.Short(err)

	// panicLine: “
	// File: "/opt/sw/parl/recover-err_test.go"
	// Line: 24
	// FuncName: "github.com/haraldrudell/parl.panicFunction"
	//”
	t.Logf("panicLine: “%s”", panicLine.Dump())

	// 	stack trace:
	// 	runtime.gopanic
	// 		/opt/homebrew/Cellar/go/1.21.3/libexec/src/runtime/panic.go:914
	// 	runtime.panicmem
	// 		/opt/homebrew/Cellar/go/1.21.3/libexec/src/runtime/panic.go:261
	// 	runtime.sigpanic
	// 		/opt/homebrew/Cellar/go/1.21.3/libexec/src/runtime/signal_unix.go:861
	// 	github.com/haraldrudell/parl.panicFunction
	// 		/opt/sw/parl/recover-err_test.go:23
	// 	github.com/haraldrudell/parl.TestRecoverErr
	// 		/opt/sw/parl/recover-err_test.go:32
	// 	testing.tRunner
	// 		/opt/homebrew/Cellar/go/1.21.3/libexec/src/testing/testing.go:1595
	// 	runtime.goexit
	// 		/opt/homebrew/Cellar/go/1.21.3/libexec/src/runtime/asm_arm64.s:1197
	// /opt/sw/parl/recover-err_test.go:43: bad error message
	t.Logf("stack trace: %s", stack)

	// error:
	// Recover from panic in runtime.gopanic:: panic:
	// 'runtime error: invalid memory address or nil pointer dereference'
	//	at runtime.gopanic()-panic.go:914
	t.Logf("error: %s", errorShort)

	// perrors.Short should detect the exact location of the panic
	var panicLineShort = panicLine.Short()
	if !strings.HasSuffix(errorShort, panicLineShort) {
		t.Errorf("perrors.Short does not end with exact panic location:\n%s\n%s",
			errorShort,
			panicLineShort,
		)
	}

	// perrors.Short should contain the message
	if !strings.Contains(errorShort, expMessage) {
		t.Errorf("perrors.Short does not contain expected error message::\n%s\n%s",
			errorShort,
			expMessage,
		)
	}
}

// panicFunction recovers a panic using [parl.RecoverErr]
//   - panicLine is the exact code line of the panic
//   - err is the error value produced by [parl.RecoverErr]
func panicFunction() (panicLine *pruntime.CodeLocation, err error) {
	defer PanicToErr(&err)

	// get exact code line and generate a nil pointer dereference panic
	var intp *int
	if false {
		var i int
		intp = &i
	}
	if panicLine = pruntime.NewCodeLocation(0); *intp != 0 {
		_ = 1
	}

	return
}
