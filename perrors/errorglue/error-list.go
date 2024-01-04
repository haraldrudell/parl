/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"

	"golang.org/x/exp/slices"
)

// ErrorList returns the list of associated errors enclosed in all error chains of err
//   - the returned slice is a list of error chains beginning with the initial err
//   - If err is nil, a nil slice is returned
//   - duplicate error values are ignored
//   - order is:
//   - — associated errors from err’s error chain, oldest first
//   - — then associated errors from other error chains, oldest found error of each chain first,
//     and errors from the last found error chain first
//   - —
//   - associated errors are independent error chains that allows a single error
//     to enclose addditional errors not part of its error chain
//   - cyclic error values are filtered out so that no error instance is returned more than once
func ErrorList(err error) (errs []error) {
	if err == nil {
		return // nil error return: nil slice
	}
	// errMap ensures that no error instance occurs more than once
	var errMap = map[error]bool{err: true}

	// process all error chains referred by err and its associated errors
	//	- begin with err
	for errorChains := []error{err}; len(errorChains) > 0; errorChains = errorChains[1:] {
		// traverse error chain
		for errorChain := errorChains[0]; errorChain != nil; errorChain = errors.Unwrap(errorChain) {

			// find any associated error not found before
			if relatedError, ok := errorChain.(RelatedError); ok {
				if associatedError := relatedError.AssociatedError(); associatedError != nil {
					if _, ok := errMap[associatedError]; !ok {

						// store associated error, newest first
						errs = append(errs, associatedError)
						// store the error chain for scanning, first found first
						errorChains = append(errorChains, associatedError)
						// store associateed error to ensure uniqueness
						errMap[associatedError] = true
					}
				}
			}
		}
	}

	// errs begins with err, then associated errors, oldest first
	errs = append(errs, err)
	slices.Reverse(errs)

	return
}
