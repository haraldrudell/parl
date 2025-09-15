/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"testing"
)

func TestIsType(t *testing.T) {
	var test string
	expected := "expected"
	wrongValue := "wrongValue"

	test = "1 err nil"
	{
		var e2 error
		t.Logf("%s", test)
		if IsType(nil, &e2) {
			t.Errorf("%s: IsType returned true FAIL", test)
		}
	}

	{ // value receiver error value
		err := valueError{s: expected}

		test = "2 value non-match"
		{
			err2 := pointerError{s: wrongValue}
			t.Logf("%s", test)
			if IsType(err, &err2) {
				t.Errorf("%s: IsType returned true FAIL", test)
			}
		}

		test = "3 value match"
		{
			err2 := valueError{s: wrongValue}
			t.Logf("%s", test)
			if !IsType(err, &err2) {
				t.Errorf("%s: IsType returned false FAIL", test)
			}
			if err2.Error() != expected {
				t.Errorf("%s: value not updated FAIL %#v exp: %q", test, err2, expected)
			}
		}
	}

	{ // value receiver error pointer
		err := &valueError{s: expected}

		test = "4 value pointer non-match"
		{
			err2 := pointerError{s: wrongValue}
			t.Logf("%s", test)
			if IsType(err, &err2) {
				t.Errorf("%s: IsType returned true FAIL", test)
			}
		}

		test = "5 value pointer match"
		{
			err2 := valueError{s: wrongValue}
			t.Logf("%s", test)
			if !IsType(err, &err2) {
				t.Errorf("%s: IsType returned false FAIL", test)
			}
			if err2.Error() != expected {
				t.Errorf("%s: value not updated FAIL %q exp: %q", test, err2.Error(), expected)
			}
		}
	}

	{ // pointer receiver
		err := &pointerError{s: expected}

		test = "6 pointer non-match"
		{ // non-match
			err2 := valueError{s: wrongValue}
			t.Logf("%s", test)
			if IsType(err, &err2) {
				t.Errorf("%s: IsType returned true FAIL", test)
			}
		}

		test = "7 pointer match"
		{
			err2 := pointerError{s: wrongValue}
			t.Logf("%s", test)
			if !IsType(err, &err2) {
				t.Errorf("%s: IsType returned false FAIL", test)
			}
			if err2.Error() != expected {
				t.Errorf("%s: value not updated FAIL %q exp: %q", test, err2.Error(), expected)
			}
		}
	}
}

func TestIsTypeErrno(t *testing.T) {

	// an error implementation with value receiver
	var err0 error = syscall.ENOENT
	var err error = fmt.Errorf("error: %w", err0)
	var actual bool
	var err2 syscall.Errno

	actual = IsType(err, &err2)
	if !actual {
		t.Errorf("errno: IsType returned false FAIL")
	}
	if err2 != err0 {
		t.Errorf("errno: IsType returned wrong value FAIL")
	}
}
func TestIsTypeSyscall(t *testing.T) {
	expected := "expected"

	// an error implementation with pointer receiver
	err0 := os.NewSyscallError(expected, errors.New(expected))
	err := fmt.Errorf("error: %w", err0)
	// pointer error implementation requires the star
	var err3 *os.SyscallError
	actual := IsType(err, &err3)
	if !actual {
		t.Errorf("syscall: IsType returned false FAIL")
	}
	if err3 == nil {
		t.Errorf("syscall: IsType returned err3 nil FAIL")
	}
	if err3 != err0 {
		t.Errorf("syscall: IsType returned wrong value %#v exp: %q#v FAIL", err3, err0)
	}
}

type valueError struct{ s string }

func (er valueError) Error() string { return er.s }

type pointerError struct{ s string }

func (er *pointerError) Error() string { return er.s }
