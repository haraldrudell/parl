package perrors

import (
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

func TestStack(t *testing.T) {
	const (
		expectedStackDepth = 1
		errMsg             = "error message"
	)
	var (
		// errorsNewError is a text error that
		// stack can be appended to
		errorsNewError = errors.New(errMsg)
		// “/opt/sw/parl/perrors/new_test.go”
		expFilename string
		err         error
	)
	if expFilename, err = expectedFileStack(); err != nil {
		t.Fatal(err.Error())
	}
	// expectedFile: /opt/sw/parl/perrors/new_test.go
	t.Logf("expectedFile: %s", expFilename)

	var (
		actualErr, errWithStack, errorNil error
		stack                             pruntime.Stack
		frames                            []pruntime.Frame
		codeLocation                      *pruntime.CodeLocation
	)

	// add stack to existing non-stack error
	errWithStack = appendStack(errorsNewError)

	// errWithStack should not be nil
	if errWithStack == nil {
		t.Fatal("FAIL err is nil")
	}

	// errWithStack.Error() should match errMsg “error message”
	if messageAct := errWithStack.Error(); messageAct != errMsg {
		t.Errorf("FAIL error message wrong: %q expected: %q",
			messageAct, errMsg,
		)
	}

	stack = errorglue.GetStackTrace(errWithStack)
	frames = stack.Frames()

	// stack: ID: 7 status: ‘running’
	// github.com/haraldrudell/parl/perrors.stackGoroutine(...)
	// 	/opt/sw/parl/perrors/stack_test.go:162
	// Parent-ID: 6 go: github.com/haraldrudell/parl/perrors.appendStack
	// 	/opt/sw/parl/perrors/stack_test.go:151
	t.Logf("stack: %s", stack)

	// number of frames should match
	if actualInt := len(frames); actualInt != expectedStackDepth {
		t.Fatalf("FAIL stack depth: %d expected: %d",
			actualInt, expectedStackDepth,
		)
	}

	// File and FuncName should not be empty
	codeLocation = frames[0].Loc()
	if !codeLocation.IsSet() || codeLocation.Line == 0 {
		t.Errorf("FAIL codeLocation.IsSet false or Line zero")
	}

	// codeLocation: github.com/haraldrudell/parl/perrors.stackGoroutine
	// /opt/sw/parl/perrors/stack_test.go:167
	t.Logf("codeLocation: %s", codeLocation)

	// filename should match expFilename
	// “/opt/sw/parl/perrors/new_test.go”
	if !strings.HasSuffix(codeLocation.File, expFilename) {
		t.Errorf("FAIL top stack frame not from this file:\n%q\n—should end:\n%q",
			codeLocation.File, expFilename,
		)
	}

	// Stack should not add stack if error already has it
	//	- Stack invoked on an error with stack should return same error
	actualErr = appendStack(errWithStack)
	if actualErr != errWithStack {
		t.Error("FAIL Stack added stack trace although it was already there")
	}

	// Stack invoked with nil should return nil
	actualErr = Stack(errorNil)
	if actualErr != nil {
		t.Error("FAIL Stack nil is not nil")
	}

	// Stackn
	actualErr = invokeStackn(errorsNewError)
	// err should not be nil
	if actualErr == nil {
		t.Error("FAIL invokeStackn returned nil")
	}

	// err.Error() should match
	if messageAct := actualErr.Error(); messageAct != errMsg {
		t.Errorf("FAIL error message wrong: %q expected: %q",
			messageAct, errMsg,
		)
	}

	stack = errorglue.GetStackTrace(actualErr)
	frames = stack.Frames()

	// stack: ID: 35 status: ‘running’
	// github.com/haraldrudell/parl/perrors.stackGoroutine({0x104d62f88?, 0x140001022e0?}, 0x14000120820?, 0x140001022f0)
	// 	/opt/sw/parl/perrors/new-errorf_test.go:213
	// Parent-ID: 34 go: github.com/haraldrudell/parl/perrors.invokeStack
	// 	/opt/sw/parl/perrors/new-errorf_test.go:203
	t.Logf("stack: %s", stack)

	// number of frames should match
	if actualInt := len(frames); actualInt != expectedStackDepth {
		t.Errorf("FAIL stack depth: %d expected: %d",
			actualInt, expectedStackDepth,
		)
	}

	codeLocation = frames[0].Loc()

	// error116.CodeLocation{
	//	 File:"/opt/sw/privates/parl/error116/new-errorf_test.go",
	// 	 Line:48,
	//	 FuncName:"github.com/haraldrudell/parl/error116.TestStack.func2"
	// }
	t.Logf("codeLocation %s", codeLocation)

	// filename should match
	if !strings.HasSuffix(codeLocation.File, expFilename) {
		t.Errorf("FAIL top stack frame not from this file: %q should end: %q",
			codeLocation.File, expFilename,
		)
	}
}

// appendStack adds a short goroutine-stack to err
//   - err: an error to add stack to
//   - errStack: the error with a short goroutine stack
func appendStack(err error) (errWithStack error) {

	// invoke the goroutine
	var ch = make(chan error)
	go stackGoroutine(err, ch)

	// await goroutine completion
	errWithStack = <-ch

	return
}

// stackGoroutine is a goroutine invoking [perrors.Stack]
//   - err: an error to assign stack to
//   - ch: a channel to send error with a short goroutine-stack
func stackGoroutine(err error, ch chan<- error) { ch <- Stack(err) }

// invokeStack invokes [perrors.Stack] in a goroutine
func invokeStackn(err error) (errWithStack error) {
	var ch = make(chan error)
	go stacknGoroutine(err, ch)
	errWithStack = <-ch

	return
}

// stackGoroutine is a goroutine invoking [perrors.Stack]
func stacknGoroutine(err error, ch chan<- error) { ch <- Stackn(err, 0) }

// expectedFile obtains the file numer for this file
//   - file:
//   - err: error from [runtime.Caller]
func expectedFileStack() (file string, err error) {
	// 0 is the caller of Caller
	// pc unitptr to get function name
	// file "/opt/sw/privates/parl/error116/stackslice_test.go"
	// line int
	if _ /*pc*/, file0, _ /*line*/, ok := runtime.Caller(0); !ok {
		err = New("runtime.Caller failed")
	} else {
		file = file0
	}

	return
}
