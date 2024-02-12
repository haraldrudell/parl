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
	"testing"

	"github.com/haraldrudell/parl/pruntime"
)

// tests a chain of errors
func TestChainString(t *testing.T) {
	//t.Error("logging on")
	// errF is a fixture with a complex error graph
	var errF = errFixture{
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
	const errorsWithStackCount = 2
	const err1FrameLength = 1

	var err, err1 error
	var messageAct, actualString string
	var errsWithStack []error
	var stack pruntime.Stack

	// err is errorStack written by FuncName goroutine
	//	- error 1 is errorStack “Prefix…”
	//	- error 2 is from [fmt.Errorf]: [fmt.wrapError]
	//	- error 3 is related error: two errors
	//	- — the associated error is errorStack “associated…”
	//	- error 4 is errorData
	//	- error 5 is errorData
	//	- error 6 is erorrStack
	//	- error 7 is errors.errorString
	err = errF.createError()

	// err should not be nil
	if err == nil {
		t.Fatal("FuncName did not update err")
	}

	// err.Error() should match
	messageAct = err.Error()
	if messageAct != errF.expectedMessage {
		t.Errorf("bad error message %q expected: %q", messageAct, errF.expectedMessage)
	}

	// stack error count should match
	errsWithStack = ErrorsWithStack(err) // error instances with stack in this error chain
	if len(errsWithStack) != errorsWithStackCount {
		t.Fatalf("FuncName did not add %d stack traces: %d", errorsWithStackCount, len(errsWithStack))
	}

	// LongFormat:
	// error-message [*errorglue.errorStack]ID: 19 IsMain: false status: running
	// github.com/haraldrudell/parl/perrors/errorglue.(*errFixture).FuncName(0x1400008e8f0)
	// 	chain-string_test.go:251
	// cre: github.com/haraldrudell/parl/perrors/errorglue.(*errFixture).Do-chain-string_test.go:215 in goroutine 18 18
	// error-message [*errors.errorString]
	err1 = errsWithStack[0]
	t.Logf("LongFormat:\n%s", ChainString(err1, LongFormat))

	// first error stack depth should match
	stack = err1.(*errorStack).StackTrace()
	if len(stack.Frames()) != err1FrameLength {
		t.Errorf("Stack length not %d: %d", err1FrameLength, len(stack.Frames()))
	}

	// err: ‘Prefix: 'error-message'’
	t.Logf("err: ‘%s’", err)

	// err DumpChain:
	// *errorglue.errorStack
	// *fmt.wrapError
	// *errorglue.relatedError
	// *errorglue.errorData
	// *errorglue.errorData
	// *errorglue.errorStack
	// *errors.errorString
	t.Logf("err DumpChain: %s", DumpChain(err))

	actualString = ChainString(err, DefaultFormat)

	// DefaultFormat: ‘Prefix: 'error-message'’
	t.Logf("DefaultFormat: ‘%s’", actualString)

	// DefaultFormat should be same as Error()
	if actualString != errF.expectedMessage {
		t.Errorf("FAIL DefaultFormat: %q expected: %q", actualString, errF.expectedMessage)
	}

	actualString = ChainString(err, ShortFormat)

	// ShortFormat:
	// ‘Prefix: 'error-message' at errorglue.(*errFixture).FuncName()-chain-string_test.go:195
	// 1[associated-error-message at errorglue.(*errFixture).FuncName()-chain-string_test.go:193]’
	t.Logf("ShortFormat: ‘%s’", actualString)

	// ShortFormat should be Error() and location
	var expected = errF.expectedMessage + " at " + errF.stack2.Frames()[0].Loc().Short()
	if !strings.HasPrefix(actualString, expected) {
		t.Errorf("FAIL ShortFormat:\n%q expected:\n%q", actualString, expected)
	}

	actualString = ChainString(err, LongFormat)

	//   error-message
	//     github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
	//       /opt/sw/privates/parl/error116/chainstring_test.go:26
	//     runtime.goexit
	//       /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
	t.Logf("LongFormat:\n%s", actualString)
}

// tests [perrrors.AppendError]
func TestAppended(t *testing.T) {
	//t.Error("logging on")
	var message1, message2 = "error1", "error2"
	var err = NewErrorStack(errors.New(message1), pruntime.NewStack(0))
	var err2 = NewRelatedError(err, NewErrorStack(errors.New(message2), pruntime.NewStack(0)))
	var prefix1 = message1 + " at errorglue."
	var contains2 = " 1[" + message2 + " at errorglue."

	var stringAct string
	_ = 1

	stringAct = ChainString(err2, ShortFormat)

	// stringAct:
	// error1 at errorglue.TestAppended()-chain-string_test.go:134
	// 1[error2 at errorglue.TestAppended()-chain-string_test.go:135]
	t.Logf("stringAct: %s", stringAct)

	// err2 shortFormat should begin with prefix1
	if !strings.HasPrefix(stringAct, prefix1) {
		t.Errorf("does not start with: %q: %q", prefix1, stringAct)
	}

	// err2 shortFormat should contain message2
	if !strings.Contains(stringAct, contains2) {
		t.Errorf("does not contain: %q: %q", contains2, stringAct)
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
	// err is an error graph
	var err = NewRelatedError(errErrorf2, relatedErr)
	// expected results for all formats
	var formatExpMap = map[CSFormat]string{
		DefaultFormat: err.Error(),
		CodeLocation:  err.Error() + " at " + stack.Frames()[0].Loc().Short(),
		ShortFormat:   err.Error() + " at " + stack.Frames()[0].Loc().Short() + " 1[related]",
		LongFormat: strings.Join([]string{
			err.Error() + " [" + reflect.TypeOf(err).String() + "]",
			errErrorf2.Error() + " [" + reflect.TypeOf(errErrorf2).String() + "]",
			errStack.Error() + " [" + reflect.TypeOf(errStack).String() + "]" + "\n" + stack.String(),
			errErrorf1.Error() + " [" + reflect.TypeOf(errErrorf1).String() + "]",
			errNew.Error() + " [" + reflect.TypeOf(errNew).String() + "]",
			relatedErr.Error() + " [" + reflect.TypeOf(relatedErr).String() + "]",
		}, "\n"),
		ShortSuffix: stack.Frames()[0].Loc().Short(),
		LongSuffix: strings.Join([]string{
			err.Error() + " [" + reflect.TypeOf(err).String() + "]",
			stack.String(),
			relatedErr.Error() + " [" + reflect.TypeOf(relatedErr).String() + "]",
		}, "\n"),
	}

	var formatAct, formatExp string
	var ok bool

	// err error-chain:
	// *errorglue.relatedError *fmt.wrapError
	// *errorglue.errorStack *fmt.wrapError *errors.errorString
	t.Logf("err error-chain: %s", DumpChain(err))

	for _, csFormat := range csFormatList {
		if formatExp, ok = formatExpMap[csFormat]; !ok {
			t.Errorf("no expected value for format: %s", csFormat)
		}
		formatAct = ChainString(err, csFormat)

		// DefaultFormat: errorf2 errorf1 new
		// ShortFormat: errorf2 errorf1 new at testing.tRunner()-testing.go:1595 1[related]
		// LongFormat: errorf2 errorf1 new [*errorglue.relatedError]
		// errorf2 errorf1 new [*fmt.wrapError]
		// errorf1 new [*errorglue.errorStack]
		// ID: 20 status: ‘running’
		// testing.tRunner(0x14000082ea0, 0x102ea5950)
		// 	/opt/homebrew/Cellar/go/1.21.7/libexec/src/testing/testing.go:1595
		// Parent-ID: 1 go: testing.(*T).Run
		// 	/opt/homebrew/Cellar/go/1.21.7/libexec/src/testing/testing.go:1648
		// errorf1 new [*fmt.wrapError]
		// new [*errors.errorString]
		// related [*errors.errorString]
		// ShortSuffix: testing.tRunner()-testing.go:1595
		// LongSuffix: errorf2 errorf1 new [*errorglue.relatedError]
		// ID: 20 status: ‘running’
		// testing.tRunner(0x14000082ea0, 0x102ea5950)
		// 	/opt/homebrew/Cellar/go/1.21.7/libexec/src/testing/testing.go:1595
		// Parent-ID: 1 go: testing.(*T).Run
		// 	/opt/homebrew/Cellar/go/1.21.7/libexec/src/testing/testing.go:1648
		// related [*errors.errorString]
		t.Logf("%s: %s", csFormat, formatAct)

		// ChainString should match
		if formatAct != formatExp {
			t.Errorf("FAIL format: %s:\n%q exp\n%q",
				csFormat, formatAct, formatExp,
			)
		}
	}
}

// shortStack retruns a short stack slice
func shortStack() (stack pruntime.Stack) { return pruntime.NewStack(2) }

// uses a goroutine to create an err fixture including
// errorStack and errorData
type errFixture struct {
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
	stack0          pruntime.Stack
	stack1          pruntime.Stack
	stack2          pruntime.Stack
}

// createError returns an error fixture
func (n *errFixture) createError() (err error) {

	// execute goroutine FuncName to end
	var ch = make(chan struct{})
	go n.FuncName(ch, &err)
	<-ch

	return
}

// goroutine that build n.err fixture
func (n *errFixture) FuncName(ch chan struct{}, errp *error) {
	defer close(ch)
	n.stack0 = pruntime.NewStack(0)
	n.stack1 = pruntime.NewStack(0)
	n.stack2 = pruntime.NewStack(0)
	// two stack traces
	// one associated error
	// a key-value string and a list string
	*errp =
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
}
