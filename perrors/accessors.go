/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package perrors

import (
	"errors"
	"reflect"

	"github.com/haraldrudell/parl/errorglue"
	"github.com/haraldrudell/parl/pruntime"
)

const e116PackFuncStackFrames = 1

// error116.ErrorData get possible string values associated with an error chain.
// list is a list of string values that were stored with an empty key, oldest first.
// keyValues are string values associated with a key string, newest key wins.
// err can be nil
func ErrorData(err error) (list []string, keyValues map[string]string) {
	for err != nil {
		if e, ok := err.(errorglue.ErrorHasData); ok {
			key, value := e.KeyValue()
			if key == "" { // for the slice
				list = append([]string{value}, list...)
			} else { // for the map
				if keyValues == nil {
					keyValues = map[string]string{key: value}
				} else if _, ok := keyValues[key]; !ok {
					keyValues[key] = value
				}
			}
		}
		err = errors.Unwrap(err)
	}
	return
}

// error116.ErrorList returns all error instances from a possible error chain.
// If err is nil an empty slice is returned.
// If err does not have associated errors, a slice of err is returned.
// otherwise, the first error of the returned slice is err followed by
// other errors oldest first.
// Cyclic error values are dropped
func ErrorList(err error) (errs []error) {
	if err == nil {
		return
	}
	err0 := err
	errMap := map[error]bool{err: true}
	for err != nil {
		if e, ok := err.(errorglue.RelatedError); ok {
			if e2 := e.AssociatedError(); e2 != nil {
				if _, ok := errMap[e2]; !ok {
					errs = append([]error{e2}, errs...)
					errMap[e2] = true
				}
			}
		}
		err = errors.Unwrap(err)
	}
	return append([]error{err0}, errs...)
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
	return PackFuncN(1)
}

func PackFuncN(skipFrames int) (packageDotFunction string) {
	if skipFrames < 0 {
		skipFrames = 0
	}
	cl := pruntime.NewCodeLocation(e116PackFuncStackFrames + skipFrames)
	packageDotFunction = cl.Name()
	if pack := cl.Package(); pack != "main" {
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
