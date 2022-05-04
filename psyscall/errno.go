/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// psyscall has functions to examine syscall.Errno errors
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

// Errno scans an error chain for a syscall.Errno type.
// Errno returns syscall.Errno 0x0 if none exists.
// Note: syscall.Errno.Error has value receiver.
// Errno checks:
//  Errno(nil) == 0 → true.
//  if errno != 0 {…
//  int(errno) provides the numeric value.
//   if errno == syscall.ENOENT…
//   fmt.Printf("%v", errno) → state not recoverable
//   fmt.Printf("0x%x", int(errno)) → 0x68
func Errno(err error) (errnoValue syscall.Errno) {
	for ; err != nil; err = errors.Unwrap(err) {
		var ok bool
		if errnoValue, ok = err.(syscall.Errno); ok {
			return // match return
		}
	}
	return // no match return
}
