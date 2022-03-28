/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package runt

import (
	"errors"
	"runtime"
)

const rtSendOnClosedChannel = "send on closed channel"

// IsSendOnClosedChannel determines if err’s chain contains the
// runtime error “send on closed channel”
func IsSendOnClosedChannel(err error) (is bool) {

	// is err runtime.Error?
	runtimeError := IsRuntimeError(err)
	if runtimeError == nil {
		return
	}

	// is it the right runtime error?
	return runtimeError.Error() == rtSendOnClosedChannel
}

// IsRuntimeError determines if err’s error chain contains a runtime.Error
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
