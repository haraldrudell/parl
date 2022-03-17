/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

import "errors"

// AllErrors obtains all error objects of all error lists that err and its error chain contains.
// If err is nil, errs is empty otherwise errs[0] == err.
// If err and its error chain has no lists, errs contain only err
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

// ErrorChainSlice returns a slice of errors from a possible error chain.
// If err is nil, an empty slice is returned.
// If err does not have an error chain, a slice of only err is returned.
// Otherwise, the slice lists each error in the chain starting with err at index 0.
func ErrorChainSlice(err error) (errs []error) {
	for err != nil {
		errs = append(errs, err)
		err = errors.Unwrap(err)
	}
	return
}

// ErrorList obtains the error list associated with err.
// is err has no list associated or the list is empty, len(errs) == 0
// err itself is not returned
func ErrorList(err error) (errs []error) {
	if errorListInstance, ok := err.(*errorList); ok {
		errs = errorListInstance.ErrorList()
	}
	return
}

func AppendError(err error, e2 error) (e error) {
	if e2 == nil {
		return err // noop return
	}
	if err == nil {
		return e2 // single error return
	}

	// check if err has list already
	if eList, ok := err.(ErrorHasList); ok {
		return eList.Append(e2)
	}
	return &errorList{ErrorChain{err}, []error{e2}}
}
