/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package error116

// errorList implements a list of errors that belong together,
// as opposed to an error chain intended for being built at
// different levels of the code.
// an error list is commonly built using error116.AppendError(err, err2) error
type errorList struct {
	ErrorChain         // errorList implements error chain, ie. rich data associated with a single error
	errs       []error // errorList has a list of additional errors
}

var _ error = &errorList{}        // errorList implements the error116.Wrapper interface, ie. is an error chain
var _ ErrorHasList = &errorList{} // errorList implements the error116.ErrorHasList and error interfaces

func (et *errorList) Append(e2 error) (e error) {
	if et == nil {
		panic(New("errorList.Append on zero value"))
	}
	if e2 == nil {
		return et // noop return
	}
	et.errs = append(et.errs, e2)
	return et
}

func (et *errorList) ErrorList() (errs []error) {
	if et != nil && len(et.errs) > 0 {
		errs = make([]error, len(et.errs))
		copy(errs, et.errs)
	}
	return
}

func (et *errorList) Count() int {
	if et == nil {
		return 0
	}
	return len(et.errs)
}
