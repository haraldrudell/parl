/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// punix has functions to examine unix.Errno errors
package punix

import (
	"errors"
	"strconv"
	"strings"

	"github.com/haraldrudell/parl"
	"golang.org/x/sys/unix"
)

// IsENOENT returns true if the root cause of err is file not found.
// Can be used with os.Open* file not found
func IsENOENT(err error) (isENOENT bool) {
	return Errno(err) == unix.ENOENT
}

// IsConnectionRefused searches the error chsin of err for unix.ECONNREFUSED
// net.Dialer errors for closed socket
func IsConnectionRefused(err error) (isConnectionRefused bool) {
	return Errno(err) == unix.ECONNREFUSED
}

// Errno scans an error chain for a unix.Errno type.
// Errno returns unix.Errno 0x0 if none exists.
// Note: unix.Errno.Error has value receiver.
// Errno checks:
//
//	Errno(nil) == 0 → true.
//	if errno != 0 {…
//	int(errno) provides the numeric value.
//	 if errno == unix.ENOENT…
//	 fmt.Printf("%v", errno) → state not recoverable
//	 fmt.Printf("0x%x", int(errno)) → 0x68
func Errno(err error) (errnoValue unix.Errno) {
	for ; err != nil; err = errors.Unwrap(err) {
		var ok bool
		if errnoValue, ok = err.(unix.Errno); ok {
			return // match return
		}
	}
	return // no match return
}

// ErrorNumberString returns the errno number as a string if
// the err error chain has a non-zero syscall.Errno error.
//   - returned string is similar to: "label: 5 0x5"
//   - if label is empty string, no label is returned
//   - if no syscall.Errno is found or it is zero, the empty string is returned
func ErrorNumberString(err error, label string) (errnoNumericString string) {
	if syscallErrno := Errno(err); syscallErrno != 0 {
		if label != "" {
			label += ": "
		}
		errnoNumericString = label + strconv.Itoa(int(syscallErrno)) + " 0x" + strconv.FormatInt(int64(syscallErrno), 16)
	}
	return
}

// ErrnoError gets the errno interpretation if the error chain does contain
// a unix.Errno type.
// if includeError is true, the error chain’s error message is prepended.
// if includeError is true and err is nil "OK" is returned
// if includeError is false or missing and no errno exists, the empty string is returned
func ErrnoError(err error, includeError ...bool) (errnoString string) {
	var isInclude bool
	if len(includeError) > 0 {
		isInclude = includeError[0]
	}

	// handle err == nil case
	if err == nil {
		if isInclude {
			return "OK"
		} else {
			return
		}
	}

	// handle includeError
	var sList []string
	if isInclude {
		sList = append(sList, err.Error())
	}

	// handle errno
	if unixErrno := Errno(err); unixErrno != 0 {
		sList = append(sList, parl.Sprintf("errno:'%s'0x%x:temporary:%t:timeout:%t",
			unixErrno.Error(),
			uint(unixErrno),
			unixErrno.Temporary(),
			unixErrno.Timeout(),
		))
	}

	return strings.Join(sList, "\x20")
}
