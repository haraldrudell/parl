/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/haraldrudell/parl/errorglue"
)

func TestStack(t *testing.T) {
	errMsg := "error message"
	expectedStackDepth := 2
	expectedFile := func() (file string) {
		// 0 is the caller of Caller
		// pc unitptr to get function name
		// file "/opt/sw/privates/parl/error116/stackslice_test.go"
		// line int
		if pc, file0, line, ok := runtime.Caller(0); !ok {
			t.Error("runtime.Caller failed")
		} else {
			_ = pc
			_ = line
			return file0
		}
		return
	}()
	var err error

	var actualString string
	var actualInt int

	getStack := func(name string, fn func(wg *sync.WaitGroup, errp *error)) (stack errorglue.StackSlice, err error) {
		var wg sync.WaitGroup
		wg.Add(1)
		go fn(&wg, &err)
		wg.Wait()
		if err == nil {
			t.Errorf("Function %s did not update err", name)
		}
		actual := err.Error()
		if actual != errMsg {
			t.Errorf("%s error message wrong: %q expected: %q", name, actual, errMsg)
		}
		// github.com/haraldrudell/parl/error116.TestStack.func2
		//   /opt/sw/privates/parl/error116/new-errorf_test.go:53
		// runtime.goexit
		//   /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
		stack = errorglue.GetStackTrace(err)
		actualInt = len(stack)
		if actualInt != expectedStackDepth {
			t.Errorf("%s stack depth: %d expected: %d", name, actualInt, expectedStackDepth)
		}
		// error116.CodeLocation{
		//	 File:"/opt/sw/privates/parl/error116/new-errorf_test.go",
		// 	 Line:48,
		//	 FuncName:"github.com/haraldrudell/parl/error116.TestStack.func2"
		// }
		codeLocation := stack[0]
		actualString = codeLocation.File
		if !strings.HasSuffix(actualString, expectedFile) {
			t.Errorf("%s top stack frame not from this file: %q should end: %q", name, actualString, expectedFile)
		}
		return
	}

	// add stack to existing non-stack error
	_, err = getStack("Add Stack", func(wg *sync.WaitGroup, errp *error) {
		*errp = Stack(errors.New(errMsg))
		wg.Done()
	})

	// do not add stack if error already has it
	_, err2 := getStack("Add Stack", func(wg *sync.WaitGroup, errp *error) {
		*errp = Stack(err)
		wg.Done()
	})
	if err != err2 {
		t.Error("Stack added stack trace although it was already there")
	}

	err = Stack(nil)
	if err != nil {
		t.Error("Stack invented an error")
	}

	// Stackn
	getStack("Stackn", func(wg *sync.WaitGroup, errp *error) {
		*errp = Stackn(errors.New(errMsg), 0)
		wg.Done()
	})
}

func TestNew(t *testing.T) {
	errPrefix := ""
	errContains := ""
	errMsg := "error message"
	expectedStackDepth := 2
	expectedFile := func() (file string) {
		// 0 is the caller of Caller
		// pc unitptr to get function name
		// file "/opt/sw/privates/parl/error116/stackslice_test.go"
		// line int
		if pc, file0, line, ok := runtime.Caller(0); !ok {
			t.Error("runtime.Caller failed")
		} else {
			_ = pc
			_ = line
			return file0
		}
		return
	}()

	getStack := func(name string, fn func(wg *sync.WaitGroup, err *error)) (err error) {
		var wg sync.WaitGroup
		wg.Add(1)
		go fn(&wg, &err)
		wg.Wait()
		if err == nil {
			t.Errorf("Function %s did not update err", name)
		}
		actual := err.Error()
		if errPrefix != "" {
			if !strings.HasPrefix(actual, errPrefix) {
				t.Errorf("%s error prefix wrong: %q expected: %q", name, actual, errPrefix)
			}
			if !strings.Contains(actual, errContains) {
				t.Errorf("%s error does not contain: %q expected: %q", name, actual, errContains)
			}
		} else if actual != errMsg {
			t.Errorf("%s error message wrong: %q expected: %q", name, actual, errMsg)
		}
		stack := errorglue.GetStackTrace(err)
		actualLen := len(stack)
		if actualLen != expectedStackDepth {
			t.Errorf("%s stack depth: %d expected: %d", name, actualLen, expectedStackDepth)
		}
		codeLocation := stack[0]
		actualFile := codeLocation.File
		if !strings.HasSuffix(actualFile, expectedFile) {
			t.Errorf("%s top stack frame not from this file: %q should end: %q", name, actualFile, expectedFile)
		}
		return
	}

	// New creates error with correct stack trace
	getStack("New", func(wg *sync.WaitGroup, err *error) {
		*err = New(errMsg)
		wg.Done()
	})

	// New add proper default message
	errPrefix = "StackNew from "
	errContains = filepath.Base(expectedFile)
	getStack("default message", func(wg *sync.WaitGroup, err *error) {
		*err = New("")
		wg.Done()
	})
}
