/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
	"strings"

	"github.com/haraldrudell/parl/pruntime"
)

/*
// AllErrors obtains all error instances references by the err error.
// This includes error instances in its error chain and those of any separate,
// associated error instances.
// If err is nil, the returned slice is empty.
// Otherwise the first error of the slice is err followed by other errors, oldest first.
// If err does not have an error chain or any associated error instances,
// the returned slice contains only err.
// Cyclic error instances are removed
func AllErrors(err error) (errs []error) {
	if err != nil {
		errs = append(errs, err)
	}

	// traverse all found errors
	errMap := map[error]bool{}
	for errsIndex := 0; errsIndex < len(errs); errsIndex++ {

		// avoid cyclic error graph
		e := errs[errsIndex]
		_, knownError := errMap[e]
		if knownError {
			continue
		}
		errMap[e] = true

		// traverse the error chain except the first entry which is the error itself
		errorChain := ErrorChainSlice(e)
		for ecIndex := 1; ecIndex < len(errorChain); ecIndex++ {

			// include a possible error list
			errs = append(errs, ErrorList(errorChain[ecIndex])...)
		}
	}
	return
}
*/

// ErrorChainSlice returns a slice of errors from a possible error chain.
// If err is nil, an empty slice is returned.
// If err does not have an error chain, a slice of only err is returned.
// Otherwise, the slice lists each error in the chain starting with err at index 0 ending with the oldest error of the chain
func ErrorChainSlice(err error) (errs []error) {
	for err != nil {
		errs = append(errs, err)
		err = errors.Unwrap(err)
	}
	return
}

// ErrorsWithStack gets all errors in the err error chain
// that has a stack trace.
// Oldest innermost stack trace is returned first.
// if not stack trace is present, the slice is empty
func ErrorsWithStack(err error) (errs []error) {
	for err != nil {
		if _, ok := err.(ErrorCallStacker); ok {
			errs = append([]error{err}, errs...)
		}
		err = errors.Unwrap(err)
	}
	return
}

// GetInnerMostStack gets the oldest stack trace in the error chain
func GetInnerMostStack(err error) (stack StackSlice) {
	var e ErrorCallStacker
	for errors.As(err, &e) {
	}
	if e != nil {
		stack = e.StackTrace()
	}
	return
}

// GetStackTrace gets the last stack trace
func GetStackTrace(err error) (stack StackSlice) {
	var e ErrorCallStacker
	if errors.As(err, &e) {
		stack = e.StackTrace()
	}
	return
}

// GetStacks gets a slice of all stack traces, oldest first
func GetStacks(err error) (stacks []StackSlice) {
	for err != nil {
		if e, ok := err.(ErrorCallStacker); ok {
			stack := e.StackTrace()
			stacks = append([]StackSlice{stack}, stacks...)
		}
		err = errors.Unwrap(err)
	}
	return
}

// DumpChain retrieves a space-separated string of
// error implementation type-names found in the error
// chain of err.
// err can be nil
//   fmt.Println(Stack(errors.New("an error")))
//   *error116.errorStack *errors.errorString
func DumpChain(err error) (typeNames string) {
	var strs []string
	for err != nil {
		strs = append(strs, fmt.Sprintf("%T", err))
		err = errors.Unwrap(err)
	}
	typeNames = strings.Join(strs, "\x20")
	return
}

// DumpGo produces a newline-separated string of
// type-names and Go-syntax found in the error
// chain of err.
// err can be nil
func DumpGo(err error) (typeNames string) {
	var strs []string
	for ; err != nil; err = errors.Unwrap(err) {
		strs = append(strs, fmt.Sprintf("%T %[1]v", err))
	}
	typeNames = strings.Join(strs, "\n")
	return
}

/*
RecoverThread is a defer function for threads.
On panic, the onError function is invoked with an error
message that contains location information
*/
func RecoverThread(label string, onError func(err error)) {
	if onError == nil {
		panic(fmt.Errorf("%s: onError func nil", pruntime.NewCodeLocation(1).PackFunc()))
	}
	if v := recover(); v != nil {
		err, ok := v.(error)
		if !ok {
			err = fmt.Errorf("Non-error value: %v", v)
		}
		onError(fmt.Errorf("%s: %w", label, err))
	}
}
