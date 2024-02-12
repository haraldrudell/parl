/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"strings"
	"sync"
	"testing"
)

func TestNewStackSlice(t *testing.T) {
	//t.Fail()

	expStackSliceLen := 2
	expFile := NewCodeLocation(0).File
	expFuncName := func(funcName string) (f string) {
		if index := strings.LastIndex(funcName, "."); index != -1 {
			f = funcName[:index+1] + "NewStackSliceThread"
		}
		return
	}(NewCodeLocation(0).FuncName)

	var stackSlice StackSlice // type StackSlice []CodeLocation
	var _ CodeLocation        // type CodeLocation struct { File Line FuncName }

	// get the stack trace
	var wg sync.WaitGroup
	wg.Add(1)
	go NewStackSliceThread(&wg, &stackSlice)
	wg.Wait()

	// the actual stackSlice:
	// github.com/haraldrudell/parl/pruntime.NewStackSliceThread
	//	  /opt/sw/monogo/parl/pruntime/stackslice_test.go:85
	// runtime.goexit
	//	  /opt/homebrew/Cellar/go/1.19.3/libexec/src/runtime/asm_arm64.s:1172
	t.Logf("stackSlice: %s", stackSlice)

	if len(stackSlice) != expStackSliceLen {
		t.Errorf("stackSlice len bad: %d exp %d", len(stackSlice), expStackSliceLen)
		t.FailNow()
	}
	if stackSlice[0].File != expFile {
		t.Errorf("stackSlice bad File: %q exp %q", stackSlice[0].File, expFile)
	}
	if stackSlice[0].FuncName != expFuncName {
		t.Errorf("stackSlice bad FuncName: %q exp %q", stackSlice[0].FuncName, expFuncName)
	}
	if stackSlice[0].Line < 1 {
		t.Errorf("stackSlice bad Line: %d exp > 0", stackSlice[0].Line)
	}
}

// get the stack slice in a thread so that we know whag the stack looks like.
// A name of a  top-level function will be present in the stack trace.
// Stack frames from test invocation are not present.
func NewStackSliceThread(wg *sync.WaitGroup, slicep *StackSlice) {
	defer wg.Done()

	// zero means skip no frames
	*slicep = NewStackSlice(0)
}
