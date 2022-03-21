/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestNewStackSlice(t *testing.T) {
	expectedStackDepth := 2

	var actualString string
	var actualInt int

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

	getStack := func(name string, fn func(wg *sync.WaitGroup, slicep *StackSlice)) (stack StackSlice) {
		var wg sync.WaitGroup
		wg.Add(1)
		go fn(&wg, &stack)
		wg.Wait()
		// github.com/haraldrudell/parl/error116.TestStack.func2
		//   /opt/sw/privates/parl/error116/new-errorf_test.go:53
		// runtime.goexit
		//   /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
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
		//t.Errorf("\n%s\n%s", actualString, expectedFile)
		if !strings.HasSuffix(actualString, expectedFile) {
			t.Errorf("%s top stack frame not from this file: %q should end: %q", name, actualString, expectedFile)
		}
		return
	}

	// add stack to existing non-stack error
	getStack("NewStackSlice", func(wg *sync.WaitGroup, slicep *StackSlice) {
		*slicep = NewStackSlice(0)
		wg.Done()
	})
}
