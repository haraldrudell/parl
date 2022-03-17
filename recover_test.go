/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"testing"
)

func TestResultPanic(t *testing.T) {

	// test recover from panic
	doFn("PANIC", // printed test name prefix
		"panic", // string used for panic
		"Recover from panic in parl.parlPanic: 'non-error value: string panic'", // expected error text
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
	return New(text)
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
