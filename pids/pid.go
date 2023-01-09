/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pids

import (
	"strconv"

	"github.com/haraldrudell/parl/ints"
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/constraints"
)

// Pid is a unique named type for process identifiers
//   - Pid implements fmt.Stringer
//   - Pid is ordered
//   - Pid has IsNonZero Int Uint32 methods
type Pid uint32

// NewPid returns a typed value process identifier
func NewPid[T constraints.Integer](pid T) (typedPid Pid, err error) {
	var u32 uint32
	if u32, err = ints.ConvertU32(pid, perrors.PackFunc()); err != nil {
		return
	}

	typedPid = Pid(u32)
	return
}

func NewPid1[T constraints.Integer](pid T) (typedPid Pid) {
	var err error
	if typedPid, err = NewPid(pid); err != nil {
		panic(err)
	}
	return
}

func (pid Pid) IsNonZero() (isValid bool) {
	return pid != 0
}

func (pid Pid) Int() (pidInt int) {
	return int(pid)
}

func (pid Pid) Uint32() (pidUint32 uint32) {
	return uint32(pid)
}

func (pid Pid) String() (s string) {
	return strconv.Itoa(int(pid))
}
