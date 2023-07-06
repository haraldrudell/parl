/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pruntime

import (
	"errors"
	"runtime"
)

// this error string appears as multiple string literals in channel.go
const rtSendOnClosedChannel = "send on closed channel"

// IsSendOnClosedChannel returns true if err’s error chain contains the
// runtime error “send on closed channel”
//   - runtime.plainError is an unexported named type of underlying type string
//   - each error occurrence has a unique runtime.plainError value
//   - runtime.plainError type have empty method RuntimeError()
//   - the [runtime.Error] interface has the RuntimeError method
//   - the Error method of runtime.Error returns the string value
func IsSendOnClosedChannel(err error) (is bool) {

	// runtimeError is any runtime.Error implementation in the err error chain
	var runtimeError = IsRuntimeError(err)
	if runtimeError == nil {
		return // not a [runtime.Error] or runtime.plainErrror return
	}

	// is it the right runtime error?
	return runtimeError.Error() == rtSendOnClosedChannel
}

// IsRuntimeError determines if err’s error chain contains a runtime.Error
//   - err1 is then a non-nil error value
//   - runtime.Error is an interface describing the unexported runtime.plainError type
func IsRuntimeError(err error) (err1 runtime.Error) {
	// errors.As cannot handle nil
	if err == nil {
		return // not an error
	}

	// is it a runtime error?
	// Go1.18: err value is a private custom type string
	// the value is effectively inaccessible
	// runtime defines an interface runtime.Error for its errors
	errors.As(err, &err1) // updates err1 if the first error in err’s chain is a runtime Error
	return
}
