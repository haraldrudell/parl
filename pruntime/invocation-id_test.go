/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"runtime/debug"
	"strings"
	"sync"
	"testing"
)

type InvocationTestType struct{}

func (tt *InvocationTestType) FuncName(wg *sync.WaitGroup, stack *string) {
	defer wg.Done()

	*stack = tt.Func2() // second-level funciton to get more levels in the stack trace
}

func (tt *InvocationTestType) Func2() (stack string) {
	return string(debug.Stack())
}

func (tt *InvocationTestType) DoInvocation() (debugStack, actual string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go tt.Invocation(&wg, &debugStack, &actual)
	wg.Wait()
	return
}

func (tt *InvocationTestType) Invocation(wg *sync.WaitGroup, debugStack, actual *string) {
	*debugStack = string(debug.Stack())
	*actual = Invocation(0)
	wg.Done()
	return
}

/*
We are generating a stack trace and want a real function name.
This means the function must be in the package scope.
Secondly, we want a short and predictable stack trace.
This means the function must be invoked as a goroutine.
*/
func TestInvocation(t *testing.T) {
	invocationIndex := 1
	debugStackIndex := 3
	var tt InvocationTestType
	debugStack, invocationResult := tt.DoInvocation()

	// get the Invocation caller’s function name:
	// github.com/haraldrudell/parl/runt.(*InvocationTestType).Invocation(0x0?, 0x0?, 0x14000104570, 0x14000104580)
	actual := strings.Split(invocationResult, "\n")[invocationIndex] // github.com/haraldrudell/parl/runt.(*InvocationTestType).Invocation(0x0?, 0x0?, 0x14000104570, 0x14000104580)

	// get our own debug.Stack value
	expected := strings.Split(debugStack, "\n")[debugStackIndex]

	if actual != expected {
		t.Errorf("Invocation failed :\n%q expected:\n%q", actual, expected)
	}
}
func TestDebugStack(t *testing.T) {
	/*
		var stack string
		var wg sync.WaitGroup
		var tt InvocationTestType
		wg.Add(1)
		go tt.FuncName(&wg, &stack)
		wg.Wait()
	*/
	/*
		stack string for go1.18: "
		goroutine 19 [running]:\n
		runtime/debug.Stack()\n
		\t/opt/homebrew/Cellar/go/1.18/libexec/src/runtime/debug/stack.go:24 +0x68\n
		github.com/haraldrudell/parl/runt.(*InvocationTestType).Func2(...)\n
		\t/opt/sw/parl/runt/invocation_test.go:25\n
		github.com/haraldrudell/parl/runt.(*InvocationTestType).FuncName(0x0?, 0x0?, 0x14000104570)\n
		\t/opt/sw/parl/runt/invocation_test.go:21 +0x50
		\ncreated by github.com/haraldrudell/parl/runt.TestInvocation\n
		\t/opt/sw/parl/runt/invocation_test.go:37 +0xb0\n"
	*/
	//t.Errorf("stack string for %s:\n%s", runtime.Version(), strconv.Quote(stack))
	/*
		the stack trace contains multiple newline-separated lines
		first line is a goroutine line
		then there are a series of 2-line stack frames, second line beginning with a tab
		first frame is debug.Stack
		last frame is the go statement creating the goroutine
		the stack trace ends with a newline
	*/
}
