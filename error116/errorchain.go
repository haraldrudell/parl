/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

// ErrorChain implements a chain of errors.
// Error chains do exist in the Go standard library but types and interfaces are not public.
// error116.ErrorChain is used as an embedded type.
// To instantiate, use a composite literal like ErrorChain{err}.
//
// error116.ErrorChainSlice returns all errors of an error chain,
// or the chain can be traversed iteratively using errors.Unwrap().
type ErrorChain struct {
	error // the wrapped error
}

var _ error = &ErrorChain{}   // ErrorChain implements the error interface
var _ Wrapper = &ErrorChain{} // ErrorChain implements the Wrapper interface

// Unwrap is a method required to make ErrorChain an error chain
// ErrorChain.Unwrap() is used by errors.Unwrap() and error116.ErrorChainSlice
func (ec *ErrorChain) Unwrap() error {
	if ec == nil {
		return nil
	}
	return ec.error
}
