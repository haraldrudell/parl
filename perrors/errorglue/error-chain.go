/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// ErrorChain implements a chain of errors.
// Error chains do exist in the Go standard library but types and interfaces are not public.
// ErrorChain is used as an embedded type.
// ErrorChain’s publics are Error() and Unwrap()
//
// ErrorChainSlice returns all errors of an error chain,
// or the chain can be traversed iteratively using errors.Unwrap()
type ErrorChain struct {
	error // the wrapped error
}

// ErrorChain behaves like an error
var _ error = &ErrorChain{}

// ErrorChain has an error chain
var _ Unwrapper = &ErrorChain{}

func newErrorChain(err error) (e2 *ErrorChain) {
	return &ErrorChain{err}
}

// Unwrap is a method required to make ErrorChain an error chain
// ErrorChain.Unwrap() is used by errors.Unwrap() and ErrorChainSlice
func (e *ErrorChain) Unwrap() (err error) {
	if e == nil {
		return
	}
	err = e.error

	return
}
