/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import "reflect"

// IsError determines if err represents error condition for all error implementations
//   - eg. unix.Errno that is uintptr
func IsError[T error](err T) (isError bool) {

	// err is interface, obtain the runtime type and check for nil runtime value
	var reflectValue = reflect.ValueOf(err)
	if !reflectValue.IsValid() {
		return // err interface has nil runtime-value return: false
	}

	// if err is not a zero-value, it is eror condition
	isError = !reflectValue.IsZero()

	return
}
