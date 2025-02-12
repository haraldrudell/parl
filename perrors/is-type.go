/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"reflect"
)

// IsType determines if the chain of err contains an error of type target.
//   - IsType is different from [errors.Is] in that IsType matches the type of err,
//     not its value.
//   - IsType is different from errors.Is in that it works for error implementations missing
//     the Is() method.
//   - IsType uses reflection.
//
// pointerToErrorValue argument is a pointer to an error implementation value, ie:
//   - if the target struct has pointer reciever, the argument type *targetStruct
//   - if the target struct has value receiver, the argument type targetStruct
func IsType(err error, pointerToErrorValue interface{}) (hadErrpType bool) {

	// ensure pointerToErrorValue is non-nil pointer
	// reflection returns nil for nil pointer
	if pointerToErrorValue == nil {
		panic(New("perrors.IsType: pointerToErrorValue nil"))
	}

	// ensure pointerToErrorValue is pointer
	pointerType := reflect.TypeOf(pointerToErrorValue)
	if pointerType.Kind() != reflect.Ptr {
		panic(New("perrors.IsType: pointerToErrorValue not pointer"))
	}

	// get the error implementation type we are looking for
	// this is what pointerToErrorValue points to
	targetType := pointerType.Elem()

	// traverse err’s error chain
	for ; err != nil; err = errors.Unwrap(err) {

		// get the type assigned to err interface
		errType := reflect.TypeOf(err)

		// check if the err type is the one we are looking for
		if errType == targetType {
			reflect.Indirect(reflect.ValueOf(pointerToErrorValue)).Set(reflect.ValueOf(err))
			return true // err match exit
		}

		// also check for what err points to
		if errType.Kind() == reflect.Ptr {
			errPointsToType := errType.Elem()
			if errPointsToType == targetType {
				reflect.Indirect(reflect.ValueOf(pointerToErrorValue)).Set(reflect.Indirect(reflect.ValueOf(err)))
				return true // *err match exit
			}
		}
	}
	return // no match exit
}
