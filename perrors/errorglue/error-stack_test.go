/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

func TestErrorStack(t *testing.T) {
	// stackSlice to use
	var stackSlice = pruntime.NewStackSlice(0)
	// error message “message”
	var message = "message"
	// at-strign preceding code locations “ at ”
	var atString = "\x20at\x20"
	// encapsulated error
	var error0 = errors.New(message)
	// a invalid format code
	var badFormat = func() (badFormat CSFormat) {
		var i = -1
		badFormat = CSFormat(i)
		return
	}()
	// extected [ShorttSuffix] output
	var shortSuffix = strings.TrimPrefix(stackSlice.Short(), atString)
	// ordered list of formats
	var formats = []CSFormat{
		DefaultFormat, ShortFormat, LongFormat, ShortSuffix, LongSuffix,
		badFormat,
	}
	// map from format to expected value
	var formatExp = map[CSFormat]string{
		DefaultFormat: error0.Error(),
		ShortFormat:   error0.Error() + stackSlice.Short(),
		LongFormat:    error0.Error() + stackSlice.String(),
		ShortSuffix:   shortSuffix,
		LongSuffix:    stackSlice.String(),
		badFormat:     "",
	}
	_ = atString
	var error1 error
	var stackSliceAct pruntime.StackSlice
	var ok bool
	var sAct, sExp string

	// ChainString() StackTrace()
	var eStack *errorStack

	// NewErrorStack should return error
	error1 = NewErrorStack(error0, stackSlice)

	// runtime type should be errorStack
	if eStack, ok = error1.(*errorStack); !ok {
		t.Fatalf("NewErrorStack not errorStack")
	}

	// StackTrace should return the slice
	stackSliceAct = eStack.StackTrace()
	if !slices.Equal(stackSliceAct, stackSlice) {
		t.Errorf("StackTrace bad\n%v exp\n%v", stackSliceAct, stackSlice)
	}

	for _, csFormat := range formats {
		if sExp, ok = formatExp[csFormat]; !ok {
			t.Errorf("no formatMap entry for format %s", csFormat)
		}
		sAct = eStack.ChainString(csFormat)

		if sAct != sExp {
			t.Errorf("ChainString %s:\n%q exp\n%q",
				csFormat,
				sAct, sExp,
			)
		}
	}
}

// ShortFormat and ShortSuffix should panic location for non-stack recover-value
func TestErrorStackPanic(t *testing.T) {
	var suffixExp string
	var atString = "\x20at\x20"
	var error0 error

	var errorRecovered error
	var suffixAct string
	var stackSlice pruntime.StackSlice

	// ChainString() StackTrace()
	var eStack *errorStack

	stackSlice, errorRecovered = getErrorStackPanic(error0)
	if errorRecovered == nil {
		panic(errors.New("errorRecovered == nil"))
	} else if stackSlice == nil {
		panic(errors.New("stackSlice == nil"))
	}

	eStack = errorRecovered.(*errorStack)

	t.Logf("stackSlice: %q", stackSlice.Short())
	suffixExp = strings.TrimPrefix(stackSlice.Short(), atString)

	suffixAct = eStack.ChainString(ShortSuffix)
	if suffixAct != suffixExp {
		t.Errorf("ChainString panic:\n%q exp\n%q", suffixAct, suffixExp)
	}
}

// ShortFormat and ShortSuffix should panic location for non-stack recover-value
func TestErrorStackPanicWithStack(t *testing.T) {
	//t.Errorf("logging on")
	var suffixExp string
	var atString = "\x20at\x20"
	var error0 = NewErrorStack(errors.New("message"), pruntime.NewStackSlice(0))

	var errorRecovered error
	var suffixAct string
	var slice pruntime.StackSlice
	var stacks []pruntime.StackSlice

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
	stacks = GetStacks(errorRecovered)
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
	t.Logf("errorRecovered oldest: %q", stacks[0].Short())
	// newest: " at errorglue.getErrorStackPanic.func1()-error-stack_test.go:156"
	t.Logf("errorRecovered newest: %q", stacks[1].Short())
	// stackSlice: " at errorglue.getErrorStackPanic()-error-stack_test.go:166"
	t.Logf("slice: %q", slice.Short())

	// get expected value from verifying slice
	suffixExp = strings.TrimPrefix(slice.Short(), atString)

	suffixAct = eStack.ChainString(ShortSuffix)
	if suffixAct != suffixExp {
		t.Errorf("ChainString panic:\n%q exp\n%q", suffixAct, suffixExp)
	}
}

func getErrorStackPanic(error0 error) (slice pruntime.StackSlice, err error) {
	defer func() {
		var stack = pruntime.NewStackSlice(0)
		var e = recover().(error)
		err = NewErrorStack(e, stack)
	}()

	if error0 == nil {
		error0 = errors.New("recover")
	}

	// NewStackSlice and panic on same line
	for slice = pruntime.NewStackSlice(0); ; panic(error0) {
	}
}
