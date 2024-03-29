/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"reflect"
	"slices"

	"github.com/haraldrudell/parl/perrors/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

const e116PackFuncStackFrames = 1

// ErrorData returns any embedded data values from err and its error chain as a list and map
//   - list contains values where key was empty, oldest first
//   - keyValues are string values associated with a key string, overwriting older values
//   - err list keyValues may be nil
func ErrorData(err error) (list []string, keyValues map[string]string) {

	// traverse the err and its error chain, newest first
	//	- only errors with data matter
	for ; err != nil; err = errors.Unwrap(err) {

		// ignore errrors without key/value pair
		var e, ok = err.(errorglue.ErrorHasData)
		if !ok {
			continue
		}

		// empty key is appended to slice
		//	- oldest value first
		var key, value = e.KeyValue()
		if key == "" { // for the slice
			list = append(list, value) // newest first
			continue
		}

		// for the map
		if keyValues == nil {
			keyValues = map[string]string{key: value}
			continue
		}
		// values are added newset first
		//	- do not overwrite newer values with older
		if _, ok := keyValues[key]; !ok {
			keyValues[key] = value
		}
	}
	slices.Reverse(list) // oldest first

	return
}

// ErrorList returns all error instances from a possible error chain.
// — If err is nil an empty slice is returned.
// — If err does not have associated errors, a slice of err, length 1, is returned.
// — otherwise, the first error of the returned slice is err followed by
//
//		other errors oldest first.
//	- Cyclic error values are dropped
func ErrorList(err error) (errs []error) {
	return errorglue.ErrorList(err)
}

// HasStack detects if the error chain already contains a stack trace
func HasStack(err error) (hasStack bool) {
	if err == nil {
		return
	}
	var e errorglue.ErrorCallStacker
	return errors.As(err, &e)
}

// IsWarning determines if an error has been flagged as a warning.
func IsWarning(err error) (isWarning bool) {
	for ; err != nil; err = errors.Unwrap(err) {
		if _, isWarning = err.(*errorglue.WarningType); isWarning {
			return // is warning
		}
	}
	return // not a warning
}

// error116.PackFunc returns the package name and function name
// of the caller:
//
//	error116.FuncName
func PackFunc() (packageDotFunction string) {
	var frames = 1 // cpunt PackFunc frame
	return PackFuncN(frames)
}

func PackFuncN(skipFrames int) (packageDotFunction string) {
	if skipFrames < 0 {
		skipFrames = 0
	}
	var cL = pruntime.NewCodeLocation(e116PackFuncStackFrames + skipFrames)
	packageDotFunction = cL.Name()
	if pack := cL.Package(); pack != "main" {
		packageDotFunction = pack + "." + packageDotFunction
	}
	return
}

func ErrpString(errp *error) (s string) {
	var err error
	if errp != nil {
		err = *errp
	}
	if err == nil {
		s = "OK"
		return
	}
	s = err.Error()
	return
}

// LongShort picks output format
//   - no error: “OK”
//   - no panic: error-message at runtime.gopanic:26
//   - panic: long format with all stack traces and values
//   - associated errors are always printed
func LongShort(err error) (message string) {
	var format errorglue.CSFormat
	if err == nil {
		format = errorglue.ShortFormat
	} else if isPanic, _, _, _ := IsPanic(err); isPanic {
		format = errorglue.LongFormat
	} else {
		format = errorglue.ShortFormat
	}
	message = errorglue.ChainString(err, format)

	return
}

// perrors.Short gets a one-line location string similar to printf %-v and ShortFormat.
// Short() does not print stack traces, data and associated errors.
// Short() does print a one-liner of the error message and a brief code location:
//
//	error-message at error116.(*csTypeName).FuncName-chainstring_test.go:26
func Short(err error) string {
	return errorglue.ChainString(err, errorglue.ShortFormat)
}

// Deferr invokes a printing function for an error pointer.
// Deferr returns the message.
// Deferr is deferrable.
// A colon and a space is appended to label.
// If *errp is nil, OK is printed.
// If errp is nil, a message is printed.
// if fn is nil, nothing is printed.
// If *errp contains an error it is printed in Short form:
//
//	label: Error message at error116.(*csTypeName).FuncName-chainstring_test.go:26
func Deferr(label string, errp *error, fn func(format string, a ...interface{})) string {
	if errp == nil {
		return "perrors.Deferr: errp nil"
	}
	if label != "" {
		label += ":\x20"
	}
	s := label + errorglue.ChainString(*errp, errorglue.ShortFormat)
	if fn != nil {
		fn(s)
	}

	return s
}

// error116.Long() gets a comprehensive string representation similar to printf %+v and LongFormat.
// ShortFormat does not print stack traces, data and associated errors.
// Long() prints full stack traces, string key-value and list values for both the error chain
// of err, and associated errors and their chains
//
//	error-message
//	  github.com/haraldrudell/parl/error116.(*csTypeName).FuncName
//	    /opt/sw/privates/parl/error116/chainstring_test.go:26
//	  runtime.goexit
//	    /opt/homebrew/Cellar/go/1.17.8/libexec/src/runtime/asm_arm64.s:1133
func Long(err error) string {
	return errorglue.ChainString(err, errorglue.LongFormat)
}

/*
IsType determines if the chain of err contains an error of type target.
IsType is different from errors.Is in that IsType matches the type of err,
not its value.
IsType is different from errors.Is in that it works for error implementations missing
the Is() method.
IsType uses reflection.
pointerToErrorValue argument is a pointer to an error implementation value, ie:

	if the target struct has pointer reciever, the argument type *targetStruct
	if the target struct has value receiver, the argument type targetStruct
*/
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

func Error0(err error) (e error) {
	for ; err != nil; err = errors.Unwrap(err) {
		e = err
	}
	return
}

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
