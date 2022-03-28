/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

// relatedError implements additional associated errors separate from the error chain.
// Associated errors allows for a function to return a single error value containing multiple error instances.
// An error chain otherwise augments a single error with additional information.
// An error list is commonly built using error116.AppendError(err, err2) error
type relatedError struct {
	ErrorChain       // errorList implements error chain, ie. rich data associated with a single error
	e          error // e is an additional error
}

var _ error = &relatedError{}        // relatedError behaves like an error
var _ Wrapper = &relatedError{}      // relatedError is an error chain
var _ RelatedError = &relatedError{} // relatedError has an associated error

func NewRelatedError(err, err2 error) (e2 error) {
	return &relatedError{*newErrorChain(err), err2}
}

func (et *relatedError) AssociatedError() (err error) {
	if et == nil {
		return
	}
	return et.e
}
