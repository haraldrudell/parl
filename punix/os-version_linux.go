//go:build linux

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import (
	"github.com/haraldrudell/parl/perrors"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

// OsVersion for linux returns version: "Linux 5.15.0-56-generic"
func OsVersion() (version string, hasVersion bool, err error) {

	var uname unix.Utsname // Unix Time-Sharing System Name
	if err = unix.Uname(&uname); perrors.Is(&err, "unix.Uname %w", err) {
		return
	}
	hasVersion = true
	version = sliceToString(uname.Sysname[:]) + "\x20" + sliceToString(uname.Release[:])

	// uname: map[
	//	- Sysname:"Linux"
	//	- Nodename:"c34z"
	//	- Release:"5.15.0-56-generic"
	//	- Version:"#62-Ubuntu SMP Tue Nov 22 19:54:14 UTC 2022"
	//	- Machine:"x86_64"
	//	- Domainname:"(none)"
	//	- ]
	// m := map[string]string{
	// 	"Sysname":    strconv.Quote(sliceToString(uname.Sysname[:])),
	// 	"Nodename":   strconv.Quote(sliceToString(uname.Nodename[:])),
	// 	"Release":    strconv.Quote(sliceToString(uname.Release[:])),
	// 	"Version":    strconv.Quote(sliceToString(uname.Version[:])),
	// 	"Machine":    strconv.Quote(sliceToString(uname.Machine[:])),
	// 	"Domainname": strconv.Quote(sliceToString(uname.Domainname[:])),
	// }
	// parl.Out("uname: %v", m)

	return
}

// sliceToString returns a string from a byte array that may contain zero-byte terminator
func sliceToString(b []byte) (s string) {
	if index := slices.Index(b, 0); index != -1 {
		b = b[:index]
	}
	s = string(b)
	return
}
