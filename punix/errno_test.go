/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import (
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

func TestIsENOENT(t *testing.T) {
	err := perrors.Errorf("err: %w", unix.ENOENT)
	if !IsENOENT(err) {
		t.Error("IsENOENT returned false")
	}

	err = perrors.New("error")
	if IsENOENT(err) {
		t.Error("IsENOENT returned true")
	}
}

func TestErrnoValues(t *testing.T) {

	// this test will only run from the command-line like:
	// go test -v -run '^TestErrnoValues$' ./punix

	// skip test if go test -v flag was not provided
	hasVFlag := false
	verboseFlag := "-test.v=true"
	for _, s := range os.Args {
		if hasVFlag = s == verboseFlag; hasVFlag {
			break
		}
	}
	if !hasVFlag {
		t.Skip()
	}

	// /var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/go-build2538455228/b001/punix.test,
	// -test.testlogfile=/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/go-build2538455228/b001/testlog.txt,
	// -test.paniconexit0,
	// -test.timeout=10m0s,
	// -test.v=true,
	// -test.run=/TestErrnoValues$
	//t.Logf("%#v", os.Args)

	var errno unix.Errno
	var _ error = errno
	var _ error = &errno
	var errnop = &errno
	_ = errnop

	// two-level indirection interface matching does not work
	// cannot use &errnop (value of type **unix.Errno) as *error value in variable declaration:
	// **unix.Errno does not implement *error (type *error is pointer to interface, not interface)
	//var _ *error = &errnop

	// invalid operation: cannot compare errno == nil
	// (mismatched types unix.Errno and untyped nil)
	//t.Logf("errno == nil: %t", errno == nil)

	// errno is compared to int, not nil
	// errno == 0 → true
	t.Logf("errno == 0 → %t", errno == 0)

	// to print properly, errno need to be cast to int
	// printing errno: %s: state not recoverable %q: "state not recoverable" %v: state not recoverable
	// printing int(errno): %d: 104 0x%x: 0x68
	errno = unix.ENOTRECOVERABLE
	t.Logf("printing errno: %%s: %s %%q: %[1]q %%v: %[1]v", errno)
	t.Logf("printing int(errno): %%d: %d 0x%%x: 0x%[1]x", int(errno))

	var e unix.Errno
	if e == 0 {
		t.Logf("if errno != 0 {…")
	}

	// %T: unix.Errno %#v: 0x68
	t.Logf("%%T: %T %%#v: %#[1]v", errno)

	t.Fail() // causes prints to be output
}

func TestErrno(t *testing.T) {
	var errno unix.Errno

	// test nil
	errno = Errno(nil)
	if errno != 0 {
		t.Errorf("punix.Errno returned non-zero")
	}

	// test non-unix.Errno
	errno = Errno(perrors.New("error"))
	if errno != 0 {
		t.Errorf("punix.Errno returned non-zero")
	}

	err := fmt.Errorf("err: %w", unix.ENOENT)

	var sList []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		sList = append(sList, reflect.TypeOf(e).String())
	}
	t.Logf("err types: %s", strings.Join(sList, "\x20"))
	t.Logf("unix.ENOENT type: %s", reflect.TypeOf(unix.ENOENT).String())

	// test non-unix.Errno
	errno = Errno(err)
	if errno == 0 {
		t.Errorf("punix.Errno returned zero")
	}
	if errno != unix.ENOENT {
		t.Errorf("punix.Errno returned wrong value: %#v", errno)
	}

}

func TestErrnoValue(t *testing.T) {
	var epermInt = 1
	var epermString = "operation not permitted"
	var epermName = "EPERM"

	var isError bool
	var errnoValue int
	var actual string

	isError = unix.EPERM != 0
	if !isError {
		t.Error("isError false")
	}
	errnoValue = int(unix.EPERM)
	if errnoValue != epermInt {
		t.Errorf("errnoValue %d exp %d", errnoValue, epermInt)
	}
	actual = fmt.Sprintf("%d", unix.EPERM)
	if actual != strconv.Itoa(epermInt) {
		t.Errorf("%%d: %q exp %q", actual, strconv.Itoa(epermInt))
	}
	actual = fmt.Sprintf("%v", unix.EPERM)
	if actual != epermString {
		t.Errorf("%%v: %q exp %q", actual, epermString)
	}
	actual = unix.ErrnoName(unix.EPERM)
	if actual != epermName {
		t.Errorf("ErrnoName: %q exp %q", actual, epermName)
	}
}

func TestErrorNumberString(t *testing.T) {
	type args struct {
		label string
		err   error
	}
	tests := []struct {
		name                   string
		args                   args
		wantErrnoNumericString string
	}{
		{"no error", args{"abc", nil}, ""},
		{"EPERM", args{"errno", unix.EPERM}, "errno: EPERM 1 0x1"},
		{"-1", args{"", unix.Errno(math.MaxUint)}, "-1 -0x1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotErrnoNumericString := ErrnoString(tt.args.label, tt.args.err); gotErrnoNumericString != tt.wantErrnoNumericString {
				t.Errorf("ErrorNumberString() = %v, want %v", gotErrnoNumericString, tt.wantErrnoNumericString)
			}
		})
	}
}
