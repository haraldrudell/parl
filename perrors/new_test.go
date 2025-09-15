/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"runtime"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

func TestNew(t *testing.T) {
	const (
		errPrefix          = "StackNew from "
		errMsg, noMsg      = "error message", ""
		expectedStackDepth = 1
	)
	var (
		fileExp, err = expectedFile()
		names        = []string{"setMessage", "defaultMessage"}
	)
	if err != nil {
		t.Fatal(err.Error())
	}

	var (
		stack            pruntime.Stack
		frames           []pruntime.Frame
		codeLocation     *pruntime.CodeLocation
		messageAct, name string
	)

	for i, message := range []string{errMsg, noMsg} {
		name = names[i]

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
//   - file:
//   - err: error from [runtime.Caller]
func expectedFile() (file string, err error) {
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

// invokeNew invokes [perrors.New] in a goroutine
//   - message: error message for New
//   - errStack: error with stack ane message
func invokeNew(message string) (errStack error) {
	var ch = make(chan error)
	go newGoroutine(message, ch)
	errStack = <-ch

	return
}

// stackGoroutine is a goroutine invoking [perrors.Stack]
func newGoroutine(message string, ch chan<- error) { ch <- New(message) }
