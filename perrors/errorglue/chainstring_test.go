/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

// we must have a package-wide type beacuse receivers can
// only be defined at the outermost level
type csTypeName struct {
	wg                                                 sync.WaitGroup
	errorMessage, errorMsg2, wrapText, expectedMessage string
	key1, value1, key2, value2                         string
	err                                                error
}

func (tn *csTypeName) Do(t *testing.T) {
	tn.wg.Add(1)
	go tn.FuncName()
	tn.wg.Wait()

	//t.Logf("LongFormat:\n%s", ChainString(tn.err, LongFormat))

	if tn.err == nil {
		t.Error("FuncName did not update err")
	}
	actual := tn.err.Error()
	if actual != tn.expectedMessage {
		t.Errorf("bad error message %q expected: %q", actual, tn.expectedMessage)
	}
	stackErrs := ErrorsWithStack(tn.err) // error instances with stack in this error chain
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

func (tn *csTypeName) FuncName() {
	// two stack traces
	// one associated error
	// a key-value string and a list string
	tn.err =
		NewErrorStack(
			fmt.Errorf(tn.wrapText,
				NewRelatedError(
					NewErrorData(
						NewErrorData(
							NewErrorStack(errors.New(tn.errorMessage), pruntime.NewStackSlice(0)),
							tn.key1, tn.value1),
						tn.key2, tn.value2),
					NewErrorStack(errors.New(tn.errorMsg2), pruntime.NewStackSlice(0)),
				)),
			pruntime.NewStackSlice(0))
	tn.wg.Done()
}

func TestChainString(t *testing.T) {
	expLines := 17
	value2Line := 5
	cst := csTypeName{
		errorMessage:    "error-message",
		errorMsg2:       "associated-error-message",
		wrapText:        "Prefix: '%w'",
		expectedMessage: "Prefix: 'error-message'",
		key1:            "key1", value1: "value1",
		key2: "", value2: "value2",
	}

	var actualString string

	cst.Do(t)

	// DefaultFormat
	// error-message
	actualString = ChainString(cst.err, DefaultFormat)
	if actualString != cst.expectedMessage {
		t.Errorf("DefaultFormat: %q expected: %q", actualString, cst.expectedMessage)
	}

	// ShortFormat
	// error1 at error116.(*csTypeName).FuncName-chainstring_test.go:26
	t.Logf("dumpChain: %s", DumpChain(cst.err))
	actualString = ChainString(cst.err, ShortFormat)
	expected := cst.expectedMessage + "\x20at\x20" + GetStacks(cst.err)[0][0].Short()
	if !strings.HasPrefix(actualString, expected) {
		t.Errorf("ShortFormat:\n%q expected:\n%q", actualString, expected)
	}

	// LongFormat
	//   error-message
	//     github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
	//       /opt/sw/privates/parl/error116/chainstring_test.go:26
	//     runtime.goexit
	//       /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
	actualString = ChainString(cst.err, LongFormat)
	//t.Errorf("LongFormat:\n%s", actualString)

	sList := strings.Split(actualString, "\n")
	if len(sList) != expLines {
		t.Errorf("LongFormat %d lines, expected: %d", len(sList), expLines)
	}
	if sList[0] != cst.expectedMessage {
		t.Errorf("LongFormat first line\n%q expected:\n%q", sList[0], cst.expectedMessage)
	}
	actualString = sList[value2Line]
	if actualString != cst.value2 {
		t.Errorf("LongFormat value2line:\n%q expected:\n%q", actualString, cst.value2)
	}

	//t.Fail()
}

func TestBuiltinErrors(t *testing.T) {
	errMsg := "error message"
	var err error
	_ = err
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		err = errors.New(errMsg)
		wg.Done()
	}()
	wg.Wait()

	// built-in errors

	// %s
	// "error message"
	//t.Error(fmt.Sprintf("%s", err))

	_ = fmt.Sprintf("%q", "abc")

	//_ = fmt.Sprintf("%q", err)

	// %v %+v %-v
	// fmt.Println (same as %v)
	// "error message"
	//t.Error(fmt.Sprintf("%v", err))
	// "error message"
	//t.Error(fmt.Sprintf("%+v", err))

	// %#v
	// &errors.errorString{s:"error message"}
	//t.Error(fmt.Sprintf("%#v", err))
}

func TestPrintNil(t *testing.T) {
	// what happens for nil error values?
	var err error
	_ = err

	// using the Error() methiod
	// panic: runtime error: invalid memory address or nil pointer dereference
	//t.Log(err.Error())

	// %v
	// <nil>
	//t.Log(fmt.Sprintf("%v", err))

	// %s
	// err: %!s(<nil>)
	//t.Log(fmt.Sprintf("err: %s", err))

	// %q
	// %!q(<nil>)
	//t.Log(fmt.Sprintf("%q", err))

	// what happens when fmt.Formatter is implemented?
	// Format is not invoked, a <nil> type string is displayed
	// {{<nil>}}
	//var re RichError // re is nil struct
	//t.Log(fmt.Sprintf("%v", re))

	// %!v(PANIC=Format method: good)
	// Format invoked with non-nil pointer value
	//err = NewRichError(nil) // err is error with value non-nil RichError pointrer with nil field
	//t.Log(fmt.Sprintf("%v", err))

	//t.Fail()
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
