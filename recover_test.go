/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"errors"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"github.com/haraldrudell/parl/pruntime"
)

func TestResultPanic(t *testing.T) {

	// test recover from panic
	doFn("PANIC", // printed test name prefix
		"panic", // string used for panic
		"Recover from panic in parl.parlPanic: “non-error value: string panic”", // expected error text
		parlPanic, // the function to invoke
		t)         // test for printing

	// test recover from error
	errX := "errX"
	doFn("ERR", errX, errX, parlError, t)

	// test success
	doFn("SUCCESS", "", "", parlSuccess, t)
}

func parlPanic(text string, fne func(error)) (err error) {
	defer Recover(Annotation(), &err, fne)
	panic(text)
}

func parlError(text string, fne func(error)) (err error) {
	defer Recover(Annotation(), &err, fne)
	return perrors.New(text)
}

func parlSuccess(text string, fne func(error)) (err error) {
	defer Recover(Annotation(), &err, fne)
	return
}

func doFn(testID, text, expectedErr string, fn func(text string, fne func(error)) (err error), t *testing.T) {
	// variables
	var errValue error
	var actualErr error
	//var actualInt int
	var actualBool bool
	//var actual string
	fne := func(e error) {
		actualBool = true
		actualErr = e
	}

	// invoke test function

	errValue = fn(text, fne)

	if expectedErr == "" {
		if actualBool {
			t.Logf("%s fne was invoked", testID)
			t.Fail()
		}
		if errValue != nil {
			t.Logf("%s returned an error: %v", testID, errValue)
			t.Fail()
		}
		if actualErr != nil {
			t.Logf("%s fne received an error: %v", testID, actualErr)
			t.FailNow()
		}
	} else {
		if !actualBool {
			t.Logf("%s fne was not invoked", testID)
			t.Fail()
		}
		if errValue == nil {
			t.Logf("%s function did not return error", testID)
			t.Fail()
		} else {
			if errValue.Error() != expectedErr {
				t.Logf("%s returned error: expected: %q actual: %s", testID, expectedErr, errValue)
				t.Fail()
			}
		}
		if actualErr == nil {
			t.Logf("%s function did not return error", testID)
			t.FailNow()
		} else {
			if actualErr.Error() != expectedErr {
				t.Logf("%s fne recived error: expected: %q actual: %q", testID, actualErr, actualErr.Error())
				t.Fail()
			}
		}
	}
}
func TestRecoverErrp(t *testing.T) {
	annotation := "annotation"
	exp := annotation + " “runtime error: invalid memory address or nil pointer dereference”"

	var err error

	func() {
		defer Recover(annotation, &err, NoOnError)
		var pt *int
		_ = *pt
	}()

	if err == nil {
		t.Error("Expected error missing")
	} else if err.Error() != exp {
		t.Errorf("bad Error: %q exp %q", err.Error(), exp)
	}
}

func TestEnsureError(t *testing.T) {
	text := "text"
	panicValue := errors.New(text)
	loc := pruntime.NewCodeLocation(0).Short()
	if index := strings.LastIndex(loc, ":"); index != -1 {
		loc = loc[:index]
	}

	err := EnsureError(panicValue)
	//var err error
	//var ok bool

	// go1.19.3: type assertion to the same type: actual already has type error (S1040)
	//if err, ok = actual.(error); !ok {

	//	t.Errorf("not error: %T %+[1]v", actual)
	//	t.FailNow()
	//}
	short := perrors.Short(err)
	if !strings.Contains(short, text) ||
		!strings.Contains(short, loc) {
		t.Errorf("bad short: %q exp %q, %q", short, text, loc)
	}
}

func TestEnsureErrorNonErr(t *testing.T) {
	panicValue := "panicValue"
	loc := pruntime.NewCodeLocation(0).Short()
	if index := strings.LastIndex(loc, ":"); index != -1 {
		loc = loc[:index]
	}

	err := EnsureError(panicValue)
	short := perrors.Short(err)
	if !strings.Contains(short, panicValue) ||
		!strings.Contains(short, loc) {
		t.Errorf("bad short: %q exp %q, %q", short, panicValue, loc)
	}
}

func TestRecoverOnError(t *testing.T) {
	// a provided annotationFixture value “annotation-fixture”
	var annotationFixture = "annotation-fixture"
	// ‘annotation-fixture “runtime error: invalid memory address or nil pointer dereference”’
	var expErrorMessage = annotationFixture + "\x20“runtime error: invalid memory address or nil pointer dereference”"
	var cL *pruntime.CodeLocation
	var errpNil *error

	var err error

	// cause a panic that is recovered and stored in err using onError function
	func() {
		defer Recover(annotationFixture, errpNil, func(e error) { err = e })

		if cL = pruntime.NewCodeLocation(0); *(*int)(nil) != 0 { // nil dereference panic
			_ = 1
		}
	}()

	// there should be a recovered error
	if err == nil {
		t.Error("Expected error missing")
		t.FailNow()
	}

	// the error message should be correct
	if err.Error() != expErrorMessage {
		t.Errorf("bad err.Error() after panic recovery:\n%q exp\n%q",
			err.Error(),
			expErrorMessage,
		)
	}

	// 221226 perrors.Short now detects panic!
	// the line provided will be the exact line of the panic,
	// not some frame from the recovering runtime.
	//	- errorglue.Indices:
	//	- recovery perrors.Short(err):
	//		"annotation-fixture 'runtime error: invalid memory address or nil pointer dereference'
	//		at parl.TestRecoverOnError.func1-recover_test.go:174"
	var errorShort = perrors.Short(err)
	var codeLineShort = cL.Short()
	if !strings.HasSuffix(errorShort, codeLineShort) {
		t.Errorf("perrors.Short(err) does not end with panic location:\n%q exp\n%q",
			errorShort,
			codeLineShort,
		)
	}

	t.Logf("recovery err.Error(): %q", err.Error())
	t.Logf("recovery perrors.Short(err): %q", perrors.Short(err))
	//t.Fail()
}

func TestAnnotation(t *testing.T) {
	exp := Sprintf("Recover from panic in %s:", pruntime.NewCodeLocation(0).PackFunc())
	actual := Annotation()
	if actual != exp {
		t.Errorf("Annotation: %q exp: %q", actual, exp)
	}
}
