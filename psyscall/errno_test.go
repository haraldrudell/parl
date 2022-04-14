/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psyscall

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestIsENOENT(t *testing.T) {
	err := perrors.Errorf("err: %w", syscall.ENOENT)
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
	// go test -v -run '^TestErrnoValues$' ./psyscall

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

	// /var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/go-build2538455228/b001/psyscall.test,
	// -test.testlogfile=/var/folders/sq/0x1_9fyn1bv907s7ypfryt1c0000gn/T/go-build2538455228/b001/testlog.txt,
	// -test.paniconexit0,
	// -test.timeout=10m0s,
	// -test.v=true,
	// -test.run=/TestErrnoValues$
	//t.Logf("%#v", os.Args)

	var errno syscall.Errno
	var _ error = errno
	var _ error = &errno
	var errnop = &errno
	_ = errnop

	// two-level indirection interface matching does not work
	// cannot use &errnop (value of type **syscall.Errno) as *error value in variable declaration:
	// **syscall.Errno does not implement *error (type *error is pointer to interface, not interface)
	//var _ *error = &errnop

	// invalid operation: cannot compare errno == nil
	// (mismatched types syscall.Errno and untyped nil)
	//t.Logf("errno == nil: %t", errno == nil)

	// errno is compared to int, not nil
	// errno == 0 → true
	t.Logf("errno == 0 → %t", errno == 0)

	// to print properly, errno need to be cast to int
	// printing errno: %s: state not recoverable %q: "state not recoverable" %v: state not recoverable
	// printing int(errno): %d: 104 0x%x: 0x68
	errno = syscall.ENOTRECOVERABLE
	t.Logf("printing errno: %%s: %s %%q: %[1]q %%v: %[1]v", errno)
	t.Logf("printing int(errno): %%d: %d 0x%%x: 0x%[1]x", int(errno))

	var e syscall.Errno
	if e == 0 {
		t.Logf("if errno != 0 {…")
	}

	// %T: syscall.Errno %#v: 0x68
	t.Logf("%%T: %T %%#v: %#[1]v", errno)

	t.Fail() // causes prints to be output
}

func TestErrno(t *testing.T) {
	var errno syscall.Errno

	// test nil
	errno = Errno(nil)
	if errno != 0 {
		t.Errorf("psyscall.Errno returned non-zero")
	}

	// test non-syscall.Errno
	errno = Errno(perrors.New("error"))
	if errno != 0 {
		t.Errorf("psyscall.Errno returned non-zero")
	}

	err := fmt.Errorf("err: %w", syscall.ENOENT)

	var sList []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		sList = append(sList, reflect.TypeOf(e).String())
	}
	t.Logf("err types: %s", strings.Join(sList, "\x20"))
	t.Logf("syscall.ENOENT type: %s", reflect.TypeOf(syscall.ENOENT).String())

	// test non-syscall.Errno
	errno = Errno(err)
	if errno == 0 {
		t.Errorf("psyscall.Errno returned zero")
	}
	if errno != syscall.ENOENT {
		t.Errorf("psyscall.Errno returned wrong value: %#v", errno)
	}
}
