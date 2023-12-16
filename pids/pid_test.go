/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pids provides a typed process identifier.
package pids

import (
	"strconv"
	"testing"

	"github.com/haraldrudell/parl/perrors"
)

func TestPid(t *testing.T) {
	var intPid = 5
	var u32 = uint32(intPid)
	var expS = strconv.Itoa(intPid)
	var badInt = -1

	var s string
	var typedPid, zeroValue Pid
	var isValid bool
	var iAct int
	var u32Act uint32
	var err error

	s = NewPid(u32).String()
	if s != expS {
		t.Errorf("String %q exp %q", s, expS)
	}

	isValid = zeroValue.IsNonZero()
	if isValid {
		t.Error("IsNonZero true")
	}

	iAct = NewPid(u32).Int()
	if iAct != intPid {
		t.Errorf("Int %d exp %d", iAct, intPid)
	}

	iAct = NewPid(u32).Int()
	if iAct != intPid {
		t.Errorf("Int %d exp %d", iAct, intPid)
	}

	u32Act = NewPid(u32).Uint32()
	if u32Act != u32 {
		t.Errorf("Uint32 %d exp %d", u32Act, u32)
	}

	typedPid, err = ConvertToPid(badInt)
	if err == nil {
		t.Error("ConvertToPid badInt missing err")
	}
	_ = typedPid

	typedPid, err = ConvertToPid(intPid)
	if err != nil {
		t.Errorf("ConvertToPid err: %s", perrors.Short(err))
	}

	if i := typedPid.Int(); i != intPid {
		t.Errorf("ConvertToPid %d exp %d", i, intPid)
	}

	if i := NewPid1(intPid).Int(); i != intPid {
		t.Errorf("NewPid1 %d exp %d", i, intPid)
	}
}
