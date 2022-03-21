/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type clTypeName struct{}

func (t *clTypeName) FuncName(wg *sync.WaitGroup,
	pc *uintptr, file *string, line *int, ok *bool, cl **CodeLocation) {
	*pc, *file, *line, *ok = runtime.Caller(0)
	*cl = NewCodeLocation(0)
	wg.Done()
}

func TestCodeLocation(t *testing.T) {

	var actualString string

	var wg sync.WaitGroup
	var pc uintptr
	var expectedFile string
	var expectedLine int
	var ok bool
	var actualCodeLocation *CodeLocation
	wg.Add(1)
	clt := clTypeName{}
	go clt.FuncName(&wg, &pc, &expectedFile, &expectedLine, &ok, &actualCodeLocation)
	wg.Wait()
	if !ok {
		t.Error("runtime.Caller failed")
	}
	// rFunc: runtime.Func is all opaque fields. methods:
	// Entry() (uintptr)
	// FileLine(uintptr) (line string, line int) "/opt/foxyboy/sw/privates/parl/mains/executable.go"
	// Name(): github.com/haraldrudell/parl/mains.(*Executable).AddErr
	var expectedFunc string
	if rFunc := runtime.FuncForPC(pc); rFunc != nil {
		expectedFunc = rFunc.Name()
	} else {
		t.Errorf("FuncForPC returned nil")
	}

	// expectedFile:
	// /opt/sw/privates/parl/error116/codelocation_test.go
	// expectedLine: 52
	// expectedFunc:
	// github.com/haraldrudell/parl/error116.TestCodeLocation.func2
	_ = expectedLine
	expectedFuncName := filepath.Base(expectedFunc)
	expectedBase := expectedFuncName
	expectedPackage := expectedFuncName
	if lastDotIndex := strings.LastIndex(expectedFuncName, "."); lastDotIndex >= 0 {
		expectedFuncName = expectedFuncName[lastDotIndex+1:]
	}
	if dotIndex := strings.Index(expectedPackage, "."); dotIndex >= 0 {
		expectedPackage = expectedPackage[:dotIndex]
	}
	expectedPackFunc := expectedPackage + "." + expectedFuncName
	expectedShortPrefix := fmt.Sprintf("%s-%s:", expectedBase, filepath.Base(expectedFile))
	expectedStringPrefix := fmt.Sprintf("%s\n\x20\x20%s:", expectedFunc, expectedFile)

	// .Name() is FuncName
	actualString = actualCodeLocation.Name()
	if actualString != expectedFuncName {
		t.Errorf("NewCodeLocation.Name() bad actual: %q expected: %q", actualString, expectedFuncName)
	}

	// .Package() is error116
	actualString = actualCodeLocation.Package()
	if actualString != expectedPackage {
		t.Errorf("NewCodeLocation.Package() bad actual: %q expected: %q", actualString, expectedPackage)
	}

	// .PackFunc() is error116.FuncName
	actualString = actualCodeLocation.PackFunc()
	if actualString != expectedPackFunc {
		t.Errorf("NewCodeLocation.PackFunc() bad actual: %q expected: %q", actualString, expectedPackFunc)
	}

	// .Base() is error116.(*TypeName).FuncName
	actualString = actualCodeLocation.Base()
	if actualString != expectedBase {
		t.Errorf("NewCodeLocation.Base() bad actual: %q expected: %q", actualString, expectedBase)
	}

	// .Short() is
	// error116.(*TypeName).FuncName-codelocation_test.go:20
	actualString = actualCodeLocation.Short()
	if !strings.HasPrefix(actualString, expectedShortPrefix) {
		t.Errorf("NewCodeLocation.Short() bad prefix: %q expected: %q", actualString, expectedBase)
	}

	// .String() is
	// github.com/haraldrudell/parl/error116.(*TypeName).FuncName\n
	//  /opt/sw/privates/parl/error116/codelocation_test.go:20
	actualString = actualCodeLocation.String()
	if !strings.HasPrefix(actualString, expectedStringPrefix) {
		t.Errorf("NewCodeLocation.String() bad prefix: %q expected: %q", actualString, expectedStringPrefix)
	}
}
