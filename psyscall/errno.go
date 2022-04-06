/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package psyscall

import (
	"errors"
	"syscall"
)

// IsENOENT returns true if the root cause of err is file not found.
// Can be used with os.Open* file not found
func IsENOENT(err error) (isENOENT bool) {
	return Errno(err) == syscall.ENOENT
}

// IsConnectionRefused searches the error chsin of err for syscall.ECONNREFUSED
// net.Dialer errors for closed socket
func IsConnectionRefused(err error) (isConnectionRefused bool) {
	return Errno(err) == syscall.ECONNREFUSED
}

// Errno scans an error chain for a syscall.Errno type, returning nil if none exists.
// int(errno) provides the numeric value.
// errno can be compared:
//  if errno == syscall.ENOENT…
// The errno numeric value can be printed:
//  fmt.Printf("0x%x", int(errno))
func Errno(err error) (errno syscall.Errno) {
	for ; err != nil; err = errors.Unwrap(err) {
		if errno1, ok := err.(syscall.Errno); ok {
			return errno1
		}
	}
	return
}
