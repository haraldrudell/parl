/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"testing"

	"golang.org/x/sys/unix"
)

func TestIsError(t *testing.T) {

	// err nil
	var err error
	if IsError(err) {
		t.Error("error(nil): true")
	}

	// err non-nil
	if !IsError(errors.New("x")) {
		t.Error("errors.New: false")
	}

	// Errno 0
	if IsError(unix.Errno(0)) {
		t.Error("unix.Errno(0): true")
	}

	// Errno non-0
	if !IsError(unix.EPERM) {
		t.Error("unix.EPERM: false")
	}

	//t.Fail()
}
