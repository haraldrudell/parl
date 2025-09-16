/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

func TestErrorStack(t *testing.T) {
	//t.Errorf("logging on")
	// stack to examine
	var stackExp = pruntime.NewStack(0)
	// at-string preceding code locations “ at ”
	var atString = "\x20at\x20"
	// encapsulated error “message”
	var error0 = errors.New("message")
	// an invalid ChainString format code
	var badFormat = func() (badFormat CSFormat) {
		var i = -1
		badFormat = CSFormat(i)
		return
	}()
	// expected [ShortSuffix] output
	var shortSuffixExp = strings.TrimPrefix(stackExp.Frames()[0].Loc().Short(), atString)
	// ordered list of formats
	var formats = []CSFormat{
		DefaultFormat, ShortFormat, LongFormat, ShortSuffix, LongSuffix,
		badFormat,
	}
	// map from format to expected value
	var formatExp map[CSFormat]string

	var ok bool
	var err error
	var stackAct pruntime.Stack
	var chainStringAct, chainStringExp string
	var _ *errorStack

	// ChainString() StackTrace()
	//	- delegated: Format() Unwrap() Error()
	var eStackAct *errorStack

	err = NewErrorStack(error0, stackExp)

	// NewErrorStack() should return runtime type errorStack
	if eStackAct, ok = err.(*errorStack); !ok {
		t.Fatalf("FAIL NewErrorStack not errorStack")
	}

	// StackTrace() should return the slice
	stackAct = eStackAct.StackTrace()
	if !slices.Equal(stackAct.Frames(), stackExp.Frames()) {
		t.Errorf("StackTrace bad\n%v exp\n%v", stackAct, stackExp)
	}

	// ChainString() should return correct string
	formatExp = map[CSFormat]string{
		DefaultFormat: eStackAct.Error(),
		ShortFormat:   eStackAct.Error() + atString + stackExp.Frames()[0].Loc().Short(),
		LongFormat:    eStackAct.Error() + " [" + reflect.TypeOf(eStackAct).String() + "]" + "\n" + stackExp.String(),
		ShortSuffix:   shortSuffixExp,
		LongSuffix:    stackExp.String(),
		badFormat:     "",
	}
	for _, csFormat := range formats {
		if chainStringExp, ok = formatExp[csFormat]; !ok {
			t.Errorf("CORRUPT no formatMap entry for format %s", csFormat)
		}

		// DefaultFormat: message
		// ShortFormat: message at errorglue.TestErrorStack()-error-stack_test.go:21
		// LongFormat: message [*errorglue.errorStack]ID: 34 IsMain: false status: running
		//     github.com/haraldrudell/parl/perrors/errorglue.TestErrorStack(0x14000120820)
		//       error-stack_test.go:21
		//     testing.tRunner(0x14000120820, 0x102b3a100)
		//       testing.go:1595
		//     cre: testing.(*T).Run-testing.go:1648 in goroutine 1 1
		// ShortSuffix: errorglue.TestErrorStack()-error-stack_test.go:21
		// LongSuffix: ID: 34 IsMain: false status: running
		// 	github.com/haraldrudell/parl/perrors/errorglue.TestErrorStack(0x14000120820)
		// 	error-stack_test.go:21
		// testing.tRunner(0x14000120820, 0x102b3a100)
		// 	testing.go:1595
		// cre: testing.(*T).Run-testing.go:1648 in goroutine 1 1
		// ?255:
		t.Logf("%s: %s", csFormat, chainStringExp)

		chainStringAct = eStackAct.ChainString(csFormat)
		if chainStringAct != chainStringExp {
			t.Errorf("FAIL ChainString %s:\n%q exp\n%q",
				csFormat,
				chainStringAct, chainStringExp,
			)
		}
	}
}

// errorStack.ChainString(ShortSuffix), and [perrors.Short], should return
// panic location and not error creation location
func TestErrorStackPanicLine(t *testing.T) {
	var atString = "\x20at\x20"

	var suffixExp string
	// errContainingPanicAct is created in recovery from a panic() invocation
	var errContainingPanicAct error
	var suffixAct string
	// stackAct is taken on the same line as a errContainingPanic panic() invocation
	var stackAct pruntime.Stack
	var noErrorBase error

	// ChainString() StackTrace()
	//	- delegated: Format() Unwrap() Error()
	var eStack *errorStack

	// get actuals
	stackAct, errContainingPanicAct = getErrorStackPanic(noErrorBase)
	if errContainingPanicAct == nil {
		panic(errors.New("errorRecovered == nil"))
	} else if stackAct == nil {
		panic(errors.New("stackSlice == nil"))
	}
	// errContainingPanicAct runtime type should be errorStack
	eStack = errContainingPanicAct.(*errorStack)

	// stackSlice: "errorglue.getErrorStackPanic()-error-stack_test.go:202"
	t.Logf("stackSlice: %s", stackAct.Frames()[0].Loc().Short())

	// ShortSuffix should match
	suffixExp = strings.TrimPrefix(stackAct.Frames()[0].Loc().Short(), atString)
	suffixAct = eStack.ChainString(ShortSuffix)
	if suffixAct != suffixExp {
		t.Errorf("FAIL ChainString panic:\n%q exp\n%q", suffixAct, suffixExp)
	}
}

// ShortFormat and ShortSuffix should panic location for non-stack recover-value
func TestErrorStackPanicWithStack(t *testing.T) {
	//t.Errorf("logging on")
	var suffixExp string
	var atString = "\x20at\x20"
	var error0 = NewErrorStack(errors.New("message"), pruntime.NewStack(0))

	var errorRecovered error
	var suffixAct string
	var slice pruntime.Stack
	var stacks []pruntime.Stack

	// ChainString() StackTrace()
	var eStack *errorStack

	slice, errorRecovered = getErrorStackPanic(error0)
	if errorRecovered == nil {
		panic(errors.New("errorRecovered == nil"))
	} else if slice == nil {
		panic(errors.New("stackSlice == nil"))
	}

	// errorRecovered should have two stacks
	//	- oldest first
	stacks = getStacks(errorRecovered)
	if len(stacks) != 2 {
		panic(errors.New("stacks not 2"))
	}
	eStack = errorRecovered.(*errorStack)

	// oldest comes from error0
	// newest comes from getErrorStackPanic
	// slice is the panic location that should be found
	//	- extracted from newest
	//	- because oldest does not have a panic

	// oldest: " at errorglue.TestErrorStackPanicWithStack()-error-stack_test.go:117"
	t.Logf("errorRecovered oldest: %q", stacks[0].Frames()[0].Loc().Short())
	// newest: " at errorglue.getErrorStackPanic.func1()-error-stack_test.go:156"
	t.Logf("errorRecovered newest: %q", stacks[1].Frames()[0].Loc().Short())
	// stackSlice: " at errorglue.getErrorStackPanic()-error-stack_test.go:166"
	t.Logf("slice: %q", slice.Frames()[0].Loc().Short())

	// get expected value from verifying slice
	suffixExp = strings.TrimPrefix(slice.Frames()[0].Loc().Short(), atString)

	suffixAct = eStack.ChainString(ShortSuffix)
	if suffixAct != suffixExp {
		t.Errorf("ChainString panic:\n%q exp\n%q", suffixAct, suffixExp)
	}
}

// getErrorStackPanic returns a stack from the same line as a panic in err
//   - error0 is an optional error used as panic argument
func getErrorStackPanic(error0 error) (stack pruntime.Stack, err error) {
	defer getErrorStackPanicRecover(&err)

	if error0 == nil {
		error0 = errors.New("recover")
	}
	// NewStack and panic on same line
	for stack = pruntime.NewStack(0); ; panic(error0) {
	}
}

// getErrorStackPanicRecover is deeferred recover function for getErrorStackPanic
func getErrorStackPanicRecover(errp *error) {
	var stack = pruntime.NewStack(0)
	var e = recover().(error)
	*errp = NewErrorStack(e, stack)
}

// getStacks gets a slice of all stack traces, oldest first
func getStacks(err error) (stacks []pruntime.Stack) {
	for err != nil {
		if e, hasStack := err.(ErrorCallStacker); hasStack {
			var stack = e.StackTrace()
			// each stack encountered is older than the previous
			// store newest first
			stacks = append(stacks, stack)
		}
		err, _, _ = Unwrap(err)
	}
	// oldest first
	slices.Reverse(stacks)

	return
}
