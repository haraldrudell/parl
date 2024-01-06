/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

func TestChainString(t *testing.T) {
	var longFormatExpectedLines = 23
	var longFormatValue2LineIndex = 8
	// cst is a fixture with a complex error graph
	var cst = errFixture{
		errorMessage:    "error-message",
		errorMsg2:       "associated-error-message",
		wrapText:        "Prefix: '%w'",
		expectedMessage: "Prefix: 'error-message'",
		key1:            "key1",
		value1:          "value1",
		// key2 is empty string
		key2:   "",
		value2: "value2",
	}
	cst.Do(t)
	var firsLineExp = cst.expectedMessage + " [" + reflect.TypeOf(cst.err).String() + "]"

	var actualString string
	var sList []string

	// cst.err.Error(): ‘Prefix: 'error-message'’
	t.Logf("cst.err.Error(): ‘%s’", cst.err.Error())
	// dumpChain:
	// *errorglue.errorStack
	// *fmt.wrapError
	// *errorglue.relatedError
	// *errorglue.errorData
	// *errorglue.errorData
	// *errorglue.errorStack
	// *errors.errorString
	t.Logf("dumpChain cst.err: %s", DumpChain(cst.err))

	// DefaultFormat should be same as Error()
	actualString = ChainString(cst.err, DefaultFormat)
	// DefaultFormat: ‘Prefix: 'error-message'’
	t.Logf("DefaultFormat: ‘%s’", actualString)
	if actualString != cst.expectedMessage {
		t.Errorf("FAIL DefaultFormat: %q expected: %q", actualString, cst.expectedMessage)
	}

	// ShortFormat should be Error() and location
	actualString = ChainString(cst.err, ShortFormat)
	// ShortFormat:
	// ‘Prefix: 'error-message' at errorglue.(*errFixture).FuncName()-chain-string_test.go:195
	// 1[associated-error-message at errorglue.(*errFixture).FuncName()-chain-string_test.go:193]’
	t.Logf("ShortFormat: ‘%s’", actualString)
	var expected = cst.expectedMessage + cst.stack2.Short()
	if !strings.HasPrefix(actualString, expected) {
		t.Errorf("FAIL ShortFormat:\n%q expected:\n%q", actualString, expected)
	}

	// LongFormat
	actualString = ChainString(cst.err, LongFormat)
	//   error-message
	//     github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
	//       /opt/sw/privates/parl/error116/chainstring_test.go:26
	//     runtime.goexit
	//       /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
	t.Logf("LongFormat:\n%s", actualString)

	// number of lines shouldmatch
	sList = strings.Split(actualString, "\n")
	if len(sList) != longFormatExpectedLines {
		t.Errorf("FAIL LongFormat %d lines, expected: %d", len(sList), longFormatExpectedLines)
	}
	// first line shouldmatch
	if sList[0] != firsLineExp {
		t.Errorf("FAIL LongFormat first line\n%q expected:\n%q", sList[0], cst.expectedMessage)
	}
	actualString = sList[longFormatValue2LineIndex]
	if actualString != cst.value2 {
		t.Errorf("FAIL LongFormat value2line:\n%q expected:\n%q", actualString, cst.value2)
	}
}

func TestAppended(t *testing.T) {
	message1 := "error1"
	prefix1 := message1 + " at errorglue."
	message2 := "error2"
	contains2 := " 1[" + message2 + " at errorglue."
	err := NewErrorStack(
		errors.New(message1),
		pruntime.NewStackSlice(0))
	err2 := NewRelatedError(err, NewErrorStack(
		errors.New(message2),
		pruntime.NewStackSlice(0)))

	var s string = ChainString(err2, ShortFormat)
	if !strings.HasPrefix(s, prefix1) {
		t.Errorf("does not start with: %q: %q", prefix1, s)
	}
	if !strings.Contains(s, contains2) {
		t.Errorf("does not contain: %q: %q", contains2, s)
	}
}

func TestChainStringList(t *testing.T) {
	var errNew = errors.New("new")
	var errErrorf1 = fmt.Errorf("errorf1 %w", errNew)
	var stack = shortStack()
	// errorStack is a rich error
	//	- does not modify error message
	var errStack = NewErrorStack(errErrorf1, stack)
	var errErrorf2 = fmt.Errorf("errorf2 %w", errStack)
	var relatedErr = errors.New("related")
	var err = NewRelatedError(errErrorf2, relatedErr)
	var formatExpMap = map[CSFormat]string{
		DefaultFormat: err.Error(),
		CodeLocation:  err.Error() + " at " + stack[0].Short(),
		ShortFormat:   err.Error() + " at " + stack[0].Short() + " 1[related]",
		LongFormat: strings.Join([]string{
			err.Error() + " [" + reflect.TypeOf(err).String() + "]",
			errErrorf2.Error() + " [" + reflect.TypeOf(errErrorf2).String() + "]",
			errStack.Error() + " [" + reflect.TypeOf(errStack).String() + "]" + stack.String(),
			errErrorf1.Error() + " [" + reflect.TypeOf(errErrorf1).String() + "]",
			errNew.Error() + " [" + reflect.TypeOf(errNew).String() + "]",
			relatedErr.Error() + " [" + reflect.TypeOf(relatedErr).String() + "]",
		}, "\n"),
		ShortSuffix: stack[0].Short(),
		LongSuffix: strings.Join([]string{
			err.Error() + " [" + reflect.TypeOf(err).String() + "]",
			stack.String(),
			relatedErr.Error() + " [" + reflect.TypeOf(relatedErr).String() + "]",
		}, "\n"),
	}

	// err error-chain:
	// *errorglue.relatedError *fmt.wrapError
	// *errorglue.errorStack *fmt.wrapError *errors.errorString
	t.Logf("err error-chain: %s", DumpChain(err))

	var formatAct, formatExp string
	var ok bool

	for _, csFormat := range csFormatList {
		if formatExp, ok = formatExpMap[csFormat]; !ok {
			t.Errorf("no expected value for format: %s", csFormat)
		}
		formatAct = ChainString(err, csFormat)

		t.Logf("%s: %s", csFormat, formatAct)

		// ChainString should match
		if formatAct != formatExp {
			t.Errorf("FAIL %s:\n%q exp\n%q",
				csFormat, formatAct, formatExp,
			)
		}
	}
}

// shortStack sends a short stack slice
func shortStack() (stack pruntime.StackSlice) {
	return pruntime.NewStackSlice(0)[:2]
}

// uses a goroutine to create an err fixture including
// errorStack and errorData
type errFixture struct {
	// err is errorStack written by FuncName goroutine
	//	- error 1 is errorStack “Prefix…”
	//	- error 2 is from [fmt.Errorf]: [fmt.wrapError]
	//	- error 3 is related error: two errors
	//	- — the associated error is errorStack “associated…”
	//	- error 4 is errorData
	//	- error 5 is errorData
	//	- error 6 is erorrStack
	//	- error 7 is errors.errorString
	err error
	// “error-message”
	errorMessage string
	// “associated-error-message”
	errorMsg2 string
	// wrapText is first associated error
	//	- “Prefix: '%w'”
	//	- used with [fmt.Errorf]
	wrapText string
	// “Prefix: 'error-message'”
	expectedMessage string
	key1            string
	value1          string
	key2            string
	value2          string
	stack0          pruntime.StackSlice
	stack1          pruntime.StackSlice
	stack2          pruntime.StackSlice
	// wg waits for the FuncName goroutine
	wg sync.WaitGroup
}

// do build err field for testing
func (n *errFixture) Do(t *testing.T) {

	// execute goroutine FuncName to end
	n.wg.Add(1)
	go n.FuncName()
	n.wg.Wait()

	//t.Logf("LongFormat:\n%s", ChainString(tn.err, LongFormat))

	if n.err == nil {
		t.Fatal("FuncName did not update err")
	}
	actual := n.err.Error()
	if actual != n.expectedMessage {
		t.Errorf("bad error message %q expected: %q", actual, n.expectedMessage)
	}
	stackErrs := ErrorsWithStack(n.err) // error instances with stack in this error chain
	if len(stackErrs) != 2 {
		t.Errorf("FuncName did not add 2 stack traces: %d", len(stackErrs))
	}
	// github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
	//   /opt/sw/privates/parl/error116/chainstring_test.go:18
	// runtime.goexit
	//   /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
	var stack pruntime.StackSlice

	t.Logf("LongFormat:\n%s", ChainString(stackErrs[0], LongFormat))

	if e, ok := stackErrs[0].(*errorStack); ok {
		stack = e.StackTrace()
	} else {
		t.Errorf("not errorStack: %T", stackErrs[0])
	}
	if len(stack) != 2 {
		t.Errorf("Stack length not 2: %d", len(stack))
	}
}

// goroutine that build n.err fixture
func (n *errFixture) FuncName() {
	n.stack0 = pruntime.NewStackSlice(0)
	n.stack1 = pruntime.NewStackSlice(0)
	n.stack2 = pruntime.NewStackSlice(0)
	// two stack traces
	// one associated error
	// a key-value string and a list string
	n.err =
		NewErrorStack(
			fmt.Errorf(n.wrapText,
				NewRelatedError(
					NewErrorData(
						NewErrorData(
							NewErrorStack(errors.New(n.errorMessage), n.stack0),
							n.key1, n.value1),
						n.key2, n.value2),
					NewErrorStack(errors.New(n.errorMsg2), n.stack1),
				)),
			n.stack2)
	n.wg.Done()
}
