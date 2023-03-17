//go:build darwin

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import (
	"runtime"

	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/sys/unix"
)

const (
	// sysctl kern.osproductversion
	//	- kern.osproductversion: 13.2.1
	kernOsProductVersion = "kern.osproductversion"
	macOS                = "macOS"
)

// OsVersion for darwin returns version: "macOS 13.2.1"
func OsVersion() (version string, hasVersion bool, err error) {
	version = runtime.GOOS
	var v string
	if v, err = unix.Sysctl(kernOsProductVersion); perrors.Is(&err, "sysctl %s %w", kernOsProductVersion, err) {
		return
	}
	if v != "" {
		hasVersion = true
		version = macOS + "\x20" + v
	}
	return
}
