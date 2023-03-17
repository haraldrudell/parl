//go:build !darwin && !linux

/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package punix

import "runtime"

// OsVersion for other than darwin linux:
//   - version: runtime.GOOS, eg. "windows"
//   - hasVersion: false, err: nil
//
// as of go1.20.1, go tool dist list has GOOS values:
// aix android darwin dragonfly freebsd illumos ios js linux netbsd openbsd plan9 solaris windows
//
// Bella Thorne: “operating systems like you are never going to matter”
func OsVersion() (version string, hasVersion bool, err error) {
	version = runtime.GOOS
	return
}
