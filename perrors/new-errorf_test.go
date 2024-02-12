/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

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
	var errMsg = "error message"
	var err0 = errors.New(errMsg)
	var expectedStackDepth = 1
	type testNo int
	const (
		addStack testNo = iota
		hasAlready
		isNil
		stackn
		LessThan
	)
	var testNames = map[testNo]string{
		addStack:   "addStack",
		hasAlready: "hasAlready",
		isNil:      "isNil",
		stackn:     "Stackn",
	}
	var fileExp, failure = expectedFile()
	if failure != "" {
		t.Fatal(failure)
	}

	// var actualInt int
	var err, err1, errorNil error
	var stack pruntime.Stack
	var frames []pruntime.Frame
	var codeLocation *pruntime.CodeLocation

	for i := testNo(0); i < LessThan; i++ {
		var name = testNames[i]
		switch i {
		case addStack:
			// add stack to existing non-stack error
			err1 = invokeStack(err0)
			err = err1
		case hasAlready:
			// do not add stack if error already has it
			err = invokeStack(err1)
			// Stack invoked on an error with stack should return same error
			if err != err1 {
				t.Error("FAIL Stack added stack trace although it was already there")
			}
			continue
		case isNil:
			// Stack invoked with nil should return nil
			err = Stack(errorNil)
			if err != nil {
				t.Error("FAIL Stack nil is not nil")
			}
			continue
		case stackn:
			// Stackn
			err = invokeStackn(err0)
		}

		// err should not be nil
		if err == nil {
			t.Errorf("FAIL test %s err is nil", name)
		}

		// err.Error() should match
		if messageAct := err.Error(); messageAct != errMsg {
			t.Errorf("FAIL %s error message wrong: %q expected: %q", name, messageAct, errMsg)
		}

		stack = errorglue.GetStackTrace(err)
		frames = stack.Frames()

		// stack: ID: 35 status: ‘running’
		// github.com/haraldrudell/parl/perrors.stackGoroutine({0x104d62f88?, 0x140001022e0?}, 0x14000120820?, 0x140001022f0)
		// 	/opt/sw/parl/perrors/new-errorf_test.go:213
		// Parent-ID: 34 go: github.com/haraldrudell/parl/perrors.invokeStack
		// 	/opt/sw/parl/perrors/new-errorf_test.go:203
		t.Logf("stack: %s", stack)

		// number of frames should match
		if actualInt := len(frames); actualInt != expectedStackDepth {
			t.Errorf("FAIL %s stack depth: %d expected: %d", name, actualInt, expectedStackDepth)
		}

		codeLocation = frames[0].Loc()

		// error116.CodeLocation{
		//	 File:"/opt/sw/privates/parl/error116/new-errorf_test.go",
		// 	 Line:48,
		//	 FuncName:"github.com/haraldrudell/parl/error116.TestStack.func2"
		// }
		t.Logf("codeLocation %s", codeLocation)

		// filename should match
		if !strings.HasSuffix(codeLocation.File, fileExp) {
			t.Errorf("FAIL %s top stack frame not from this file: %q should end: %q", name, codeLocation.File, fileExp)
		}
	}
}

func TestNew(t *testing.T) {
	var errPrefix = "StackNew from "
	// var errContains = ""
	var errMsg, noMsg = "error message", ""
	var expectedStackDepth = 1
	var fileExp, failure = expectedFile()
	if failure != "" {
		t.Fatal(failure)
	}
	var names = []string{"setMessage", "defaultMessage"}

	var err error
	var stack pruntime.Stack
	var frames []pruntime.Frame
	var codeLocation *pruntime.CodeLocation
	var messageAct string

	for i, message := range []string{errMsg, noMsg} {
		var name = names[i]

		// invoke [perrors.New]
		err = invokeNew(message)

		// err should not be nil
		if err == nil {
			t.Errorf("FAIL test %s err is nil", name)
		}

		messageAct = err.Error()
		if i == 0 {
			// err.Error() should match
			if messageAct != errMsg {
				t.Errorf("FAIL %s error message wrong: %q expected: %q", name, messageAct, errMsg)
			}
			// New add proper default message
		} else if !strings.HasPrefix(messageAct, errPrefix) {
			t.Errorf("FAIL %s error message wrong: %q expected prefix: %q", name, messageAct, errPrefix)
		}

		stack = errorglue.GetStackTrace(err)
		frames = stack.Frames()

		// stack: ID: 35 status: ‘running’
		// github.com/haraldrudell/parl/perrors.newGoroutine({0x1045a2f0e?, 0x10450a934?}, 0x140001209c0?, 0x140001022f0)
		// 	/opt/sw/parl/perrors/new-errorf_test.go:241
		// Parent-ID: 34 go: github.com/haraldrudell/parl/perrors.invokeNew
		// 	/opt/sw/parl/perrors/new-errorf_test.go:231
		t.Logf("stack: %s", stack)

		// number of frames should match
		if actualInt := len(frames); actualInt != expectedStackDepth {
			t.Errorf("FAIL %s stack depth: %d expected: %d", name, actualInt, expectedStackDepth)
		}

		codeLocation = frames[0].Loc()

		// codeLocation github.com/haraldrudell/parl/perrors.newGoroutine
		// /opt/sw/parl/perrors/new-errorf_test.go:241
		t.Logf("codeLocation %s", codeLocation)

		// filename should match
		if !strings.HasSuffix(codeLocation.File, fileExp) {
			t.Errorf("FAIL %s top stack frame not from this file: %q should end: %q", name, codeLocation.File, fileExp)
		}
	}
}

// expectedFile obtains the file numer for this file
func expectedFile() (file, failure string) {
	// 0 is the caller of Caller
	// pc unitptr to get function name
	// file "/opt/sw/privates/parl/error116/stackslice_test.go"
	// line int
	if pc, file0, line, ok := runtime.Caller(0); !ok {
		failure = "runtime.Caller failed"
	} else {
		_ = pc
		_ = line
		file = file0
	}
	return
}

// invokeStack invokes [perrors.Stack] in a goroutine
func invokeStack(err error) (errStack error) {
	var ch = make(chan struct{})
	go stackGoroutine(err, ch, &errStack)
	<-ch

	return
}

// stackGoroutine is a goroutine invoking [perrors.Stack]
func stackGoroutine(err error, ch chan struct{}, errp *error) {
	defer close(ch)

	*errp = Stack(err)
}

// invokeStack invokes [perrors.Stack] in a goroutine
func invokeStackn(err error) (errStack error) {
	var ch = make(chan struct{})
	go stacknGoroutine(err, ch, &errStack)
	<-ch

	return
}

// stackGoroutine is a goroutine invoking [perrors.Stack]
func stacknGoroutine(err error, ch chan struct{}, errp *error) {
	defer close(ch)

	*errp = Stackn(err, 0)
}

// invokeNew invokes [perrors.New] in a goroutine
func invokeNew(message string) (errStack error) {
	var ch = make(chan struct{})
	go newGoroutine(message, ch, &errStack)
	<-ch

	return
}

// stackGoroutine is a goroutine invoking [perrors.Stack]
func newGoroutine(message string, ch chan struct{}, errp *error) {
	defer close(ch)

	*errp = New(message)
}
