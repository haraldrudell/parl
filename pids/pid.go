/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// Package pids provides a typed process identifier.
package pids

import (
	"strconv"

	"github.com/haraldrudell/parl/ints"
	"github.com/haraldrudell/parl/pruntime"
	"golang.org/x/exp/constraints"
)

// Pid is a unique named type for process identifiers
//   - Pid implements [fmt.Stringer]
//   - Pid is [constraints.Ordered]
//   - Pid has [Pid.IsNonZero] [Pid.Int] [Pid.Uint32] methods
type Pid uint32

// NewPid returns a process identifier based on a 32-bit integer
func NewPid(u32 uint32) (pid Pid) { return Pid(u32) }

// NewPid1 returns a typed value process identifier panicking on error
func NewPid1[T constraints.Integer](pid T) (typedPid Pid) {
	var err error
	if typedPid, err = ConvertToPid(pid); err != nil {
		panic(err)
	}
	return
}

// ConvertToPid returns a typed value process identifier from any Integer type
func ConvertToPid[T constraints.Integer](pid T) (typedPid Pid, err error) {
	var u32 uint32
	if u32, err = ints.Unsigned[uint32](pid, pruntime.PackFunc()); err != nil {
		return
	}
	typedPid = Pid(u32)

	return
}

// IsNonZero returns whether trhe process identifier contains a valid process ID
func (pid Pid) IsNonZero() (isValid bool) { return pid != 0 }

// Int converts a process identifier to a platform-specific sized int
func (pid Pid) Int() (pidInt int) { return int(pid) }

// Uint32 converts a process identifier to a 32-bit unsigned integer
func (pid Pid) Uint32() (pidUint32 uint32) { return uint32(pid) }

func (pid Pid) String() (s string) { return strconv.Itoa(int(pid)) }
