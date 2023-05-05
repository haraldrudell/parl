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
//   - if no errno value exists, unix.Errno 0x0 is returned
//   - unlike most error implementations, unix.Errno type is uintptr
//   - to check for error condition of a unix.Errno error: if errno != 0 {…
//   - to obtain errno number: int(errno)
//   - to print errno number: fmt.Sprintf("%d", errno) → "1"
//   - to print errno message: fmt.Sprintf("%v", unix.EPERM) → "operation not permitted"
//   - to obtain errno name: unix.ErrnoName(unix.EPERM) → "EPERM"
//   - to print hexadecimal errno:
//
// Note: unix.Errno.Error has value receiver.
// Errno checks:
//
//	Errno(nil) == 0 → true.
//	if errno != 0 {…
//	int(errno) // numeric value
//	 if errno == unix.ENOENT…
//	 fmt.Sprintf("%d", unix.EPERM) → "1"
//	 fmt.Printf("%v", errno) → "state not recoverable"
//	 unix.ErrnoName(unix.EPERM) → "EPERM"
//	 var i, s = int(errno), ""
//	 if i < 0 { i = -i; s = "-" }
//	 fmt.Printf("%s0x%x", s, i) → 0x68
func Errno(err error) (errnoValue unix.Errno) {
	for ; err != nil; err = errors.Unwrap(err) {
		var ok bool
		if errnoValue, ok = err.(unix.Errno); ok {
			return // match return
		}
	}
	return // no match return
}

// ErrnoString returns the errno number as a string if
// the err error chain has a non-zero syscall.Errno error.
//   - if label is empty string, no label is returned
//   - if no syscall.Errno is found or it is zero, the empty string is returned
//   - ErrnoString("errno", nil) → ""
//   - ErrnoString("errno", unix.EPERM) → "errno: EPERM 1 0x1"
//   - ErrnoString("errno", unix.Errno(math.MaxUint)) → "-1 -0x1"
func ErrnoString(label string, err error) (errnoNumericString string) {
	var unixErrno = Errno(err)
	if unixErrno == 0 {
		return // no errno error return: ""
	}

	if label != "" {
		errnoNumericString = label + ":\x20"
	}

	if name := unix.ErrnoName(unixErrno); name != "" {
		errnoNumericString += name + "\x20"
	}

	var errno = int(unixErrno)
	errnoNumericString += strconv.Itoa(errno) + "\x20"

	var hexErrno = int64(errno)
	var sign string
	if hexErrno < 0 {
		sign = "-"
		hexErrno = -hexErrno
	}
	errnoNumericString += sign + "0x" + strconv.FormatInt(hexErrno, 16)

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
