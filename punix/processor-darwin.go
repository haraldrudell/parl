//go:build darwin

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import (
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

const (
	machdepCpuBrandString = "machdep.cpu.brand_string"
)

// Processor returns human readable string like "Apple M1 Max"
func Processor() (model string, err error) {
	if model, err = unix.Sysctl(machdepCpuBrandString); perrors.Is(&err, "sysctl %s %w", machdepCpuBrandString, err) {
		return
	}
	return
}
