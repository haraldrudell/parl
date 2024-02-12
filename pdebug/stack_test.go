/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pdebug

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

func TestStack(t *testing.T) {
	var threadID, expStatus = func() (threadID parl.ThreadID, status parl.ThreadStatus) {
		// "goroutine 34 [running]:"
		line := strings.Split(string(debug.Stack()), "\n")[0]
		values := strings.Split(line, "\x20")
		if u64, err := strconv.ParseUint(values[1], 10, 64); err != nil {
			panic(err)
		} else {
			threadID = parl.ThreadID(u64)
		}
		status = parl.ThreadStatus(values[2][1 : len(values[2])-2])
		return
	}()
	var expCreatorPrefix = "testing.(*T)."
	var expShorts0 = fmt.Sprintf("Thread ID: %s", threadID)

	var expGoFunctionFuncName, expGoFunctionFileLine = func() (funcName, fileLine string) {
		//	- -5: goFunction function
		// - -4: goFunction file-line
		// - -3: cre: function
		// - -2: creator file-line
		// - -1:  last line is empty
		lines := strings.Split(string(debug.Stack()), "\n")
		if len(lines) < 5 {
			panic(perrors.ErrorfPF("too few lines in stack trace: %d need >=5", len(lines)))
		}
		funcLine := lines[len(lines)-5] // "testing.tRunner(0x1400011cb60, 0x1011810a8)"
		packagePath, packageName, typePath, funcName := pruntime.SplitAbsoluteFunctionName(funcLine)
		if argIndex := strings.Index(funcName, "("); argIndex != -1 {
			funcName = funcName[:argIndex]
		}
		// "testing.tRunner"
		funcName = packagePath + packageName + "." + typePath + funcName
		// "\t/opt/homebrew/Cellar/go/1.20.4/libexec/src/testing/testing.go:1576 +0x10c"
		fileLine = lines[len(lines)-4]
		if len(fileLine) > 0 {
			fileLine = fileLine[1:]
		}
		if endIndex := strings.Index(fileLine, "\x20"); endIndex != -1 {
			fileLine = fileLine[:endIndex]
		}
		return
	}()

	var frames []pruntime.Frame
	var actualS string
	var actualSL []string

	var stack, cL = NewStack(0), pruntime.NewCodeLocation(0)
	if stack == nil {
		t.Fatalf("NewStack nil return")
	}

	// ID() Status() IsMain() Frames() Creator() Shorts()
	var _ parl.Stack
	if stack.ID() != threadID {
		t.Errorf("stack.ID: %s exp %s", stack.ID(), threadID)
	}
	if stack.Status() != expStatus {
		t.Errorf("stack.Status: %s exp %s", stack.Status(), expStatus)
	}
	if stack.IsMain() {
		t.Errorf("stack.IsMain true: stack:\n%s\n", stack.String())
	}
	frames = stack.Frames()
	if len(frames) < 1 {
		t.Error("stack.Frames empty")
	}
	if frames[0].Loc().Short() != cL.Short() {
		t.Errorf("stack.Frames[0]: %s exp %s", frames[0].Loc().Short(), cL.Short())
	}
	// File: "/opt/homebrew/Cellar/go/1.20.4/libexec/src/testing/testing.go" Line: 1576 FuncName: "testing.tRunner"
	t.Logf(stack.GoFunction().Dump())
	// expGoFunctionFuncName "testing.tRunner"
	t.Logf("expGoFunctionFuncName %q", expGoFunctionFuncName)
	// "/opt/homebrew/Cellar/go/1.20.4/libexec/src/testing/testing.go:1576"
	t.Logf("expGoFunctionFileLine %q", expGoFunctionFileLine)
	if stack.GoFunction().PackFunc() != expGoFunctionFuncName {
		t.Errorf("stack.GoFunction PackFunc: %q expPrefix %q", stack.GoFunction().PackFunc(), expGoFunctionFuncName)
	}
	actualS = stack.GoFunction().File + ":" + strconv.Itoa(stack.GoFunction().Line)
	if actualS != expGoFunctionFileLine {
		t.Errorf("stack.GoFunction FuncLine: %q expPrefix %q", stack.GoFunction().FuncLine(), expGoFunctionFileLine)
	}
	var creator, _, _ = stack.Creator()
	actualS = creator.Short()
	if !strings.HasPrefix(actualS, expCreatorPrefix) {
		t.Errorf("stack.Creator: %q expPrefix %q", actualS, expCreatorPrefix)
	}
	actualSL = strings.Split(stack.Shorts(""), "\n")
	if actualSL[0] != expShorts0 {
		t.Errorf("stack.Shorts 0: %q exp %q", actualSL[0], expShorts0)
	}
	if actualSL[1] != cL.Short() {
		t.Errorf("stack.Shorts 1: %q exp %q", actualSL[1], cL.Short())
	}
}

func TestStackString(t *testing.T) {
	var creator, goFunction, frame0cL pruntime.CodeLocation
	var stack parl.Stack
	var threadID, parentID parl.ThreadID
	var threadStatus parl.ThreadStatus

	// populate variables above with predictable values
	var tt = T{
		creator:      &creator,
		goFunction:   &goFunction,
		stack:        &stack,
		frame0:       &frame0cL,
		threadID:     &threadID,
		threadStatus: &threadStatus,
		parentID:     &parentID,
	}
	tt.wg.Add(1)
	go stackGenerator(&tt, pruntime.NewCodeLocation(0))
	tt.wg.Wait()

	// assemble expected values

	// The stack is created by T.stack2
	//	- 1 status line: ID: 35 IsMain: false status: running
	//	- 3x2 frames: pdebug.NewStack
	//	- 1 creator line: cre: github.com/haraldrudell/parl/pdebug.TestStackString-stack_test.go:161
	var expLines = 1 + 2*2 + 2 // 6: 0–5: ID, 2x2code locations, cre
	var expLine1 = fmt.Sprintf("ID: %s status: ‘%s’", threadID, threadStatus)
	var expLine3 = "\x20\x20" + frame0cL.File + ":" + strconv.Itoa(frame0cL.Line)
	var expLine5 = "\x20\x20" + goFunction.File + ":" + strconv.Itoa(goFunction.Line)
	var expLine6 = strings.Split(fmt.Sprintf("Parent-ID: %d go: %s", parentID, creator), "\n")[0]

	var actualSL []string

	// STACK OBTAINED:
	// ID: 37 status: ‘running’
	// github.com/haraldrudell/parl/pdebug.stack2(0x1400014aa80, 0x14000114cc0, 0x14000114c90)
	// 	/opt/sw/parl/pdebug/stack_test.go:137
	// github.com/haraldrudell/parl/pdebug.stackGenerator(0x14000104ea0?, 0x100abbfd0?)
	// 	/opt/sw/parl/pdebug/stack_test.go:132
	// Parent-ID: 36 go: github.com/haraldrudell/parl/pdebug.TestStackString
	// 	/opt/sw/parl/pdebug/stack_test.go:174
	// —
	t.Logf("STACK OBTAINED:\n%s\n—", stack)

	actualSL = strings.Split(stack.String(), "\n")
	// length: 6: 0…5
	if len(actualSL) != expLines {
		t.Fatalf("FAIL stack.String lines: %d exp %d", len(actualSL), expLines)
	}
	// ID: 35 IsMain: false status: running
	if actualSL[0] != expLine1 {
		t.Errorf("FAIL stack.String 1: %q exp %q", actualSL[0], expLine1)
	}
	// github.com/haraldrudell/parl/pdebug.stack2(0x1400011ea80, 0x1400010f080, 0x1400010f050)
	if !strings.HasPrefix(actualSL[1], frame0cL.FuncName) {
		t.Errorf("FAIL stack.String frame 0: %q expPrefix %q", actualSL[1], frame0cL.FuncName)
	}
	// stack_test.go:86
	if actualSL[2] != expLine3 {
		t.Errorf("FAIL stack.String frame 1: %q exp %q", actualSL[2], expLine3)
	}
	// github.com/haraldrudell/parl/pdebug.stackGenerator(0x14000003ba0?, 0x102d05020?)
	if !strings.HasPrefix(actualSL[3], goFunction.FuncName) {
		t.Errorf("FAIL stack.String frame 2: %q expPrefix %q", actualSL[3], goFunction.FuncName)
	}
	// stack_test.go:82
	if actualSL[4] != expLine5 {
		t.Errorf("FAIL stack.String frame 3: %q exp %q", actualSL[4], expLine5)
	}
	// cre: github.com/haraldrudell/parl/pdebug.TestStackString-stack_test.go:119
	if !strings.HasPrefix(actualSL[5], expLine6) {
		t.Errorf("FAIL stack.String Creator:\n%q expPrefix\n%q", actualSL[5], expLine6)
	}
}

func TestParseFirstStackLine(t *testing.T) {
	input := []byte("goroutine 19 [running]:\ngarbage")
	expID := parl.ThreadID(19)
	expStatus := parl.ThreadStatus("running")

	ID, status, err := ParseFirstLine(input)
	if err != nil {
		t.Errorf("FAIL ParseFirstStackLine err: %v", err)
	}
	if ID != expID {
		t.Errorf("FAIL ID: %q exp: %q", ID, expID)
	}
	if status != expStatus {
		t.Errorf("FAIL status: %q exp: %q", status, expStatus)
	}
}

// END OF TEST
// END OF TEST
// END OF TEST

type T struct {
	creator, goFunction, frame0 *pruntime.CodeLocation
	threadID                    *parl.ThreadID
	threadStatus                *parl.ThreadStatus
	stack                       *parl.Stack
	parentID                    *parl.ThreadID
	wg                          sync.WaitGroup
}

func stackGenerator(t *T, c *pruntime.CodeLocation) { stack2(t, pruntime.NewCodeLocation(0), c) }

func stack2(t *T, goFunction, creator *pruntime.CodeLocation) {
	defer t.wg.Done()

	var stack, cL = NewStack(0), pruntime.NewCodeLocation(0)
	*t.creator = *creator
	*t.goFunction = *goFunction
	*t.stack = stack
	*t.frame0 = *cL
	threadID, status := func() (threadID parl.ThreadID, status parl.ThreadStatus) {
		// "goroutine 34 [running]:"
		line := strings.Split(string(debug.Stack()), "\n")[0]
		values := strings.Split(line, "\x20")
		if u64, err := strconv.ParseUint(values[1], 10, 64); err != nil {
			panic(err)
		} else {
			threadID = parl.ThreadID(u64)
		}
		status = parl.ThreadStatus(values[2][1 : len(values[2])-2])
		return
	}()
	*t.threadID = threadID
	*t.threadStatus = status
	_, *t.parentID, _ = stack.Creator()
}

// single-step through constructor
// func TestStack0(t *testing.T) {
// 	var stack parl.Stack = NewStack(0)
// 	_ = stack
// }
