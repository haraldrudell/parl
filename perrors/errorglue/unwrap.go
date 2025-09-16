/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import "errors"

// Unwrap unwraps one step of an error-chain,
// returning a node from a directed graph of error values
//   - err: error to follow, may be nil
//   - nextError: the next error in the main error-chain,
//     typically from an [fmt.Errorf] expression
//   - nextError nil: there are no more errors
//   - joinErrors: errors 1… from an [error.Join] expression.
//     May be nil or empty
//   - associatedError: any error from an [perrors.AppendError] expression.
//     May be nil
//   - —
//   - all error-chain traversal of Parl should use this function
//   - the Go standrad library do not offer a traversal function for
//     [errors.Join] values
func Unwrap(err error) (nextError error, joinedErrors []error, associatedError error) {

	if relatedError, hasAssociatedError := err.(RelatedError); hasAssociatedError {
		associatedError = relatedError.AssociatedError()
	}

	switch wrapper := err.(type) {
	case Unwrapper:
		nextError = wrapper.Unwrap()
	case JoinUnwrapper:
		joinedErrors = wrapper.Unwrap()
		nextError = joinedErrors[0]
		joinedErrors = joinedErrors[1:]
	}

	return
}

// errors.Join is the only source of JoinUnwrapper and always
// returns pointer implementation
var _ = errors.Join
