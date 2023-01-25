/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package recover

import "errors"

// ErrNilValue indicates that a panic was caused by a value such as a
// function argument was nil that cannot be nil.
//
// Usage:
//
//	if errors.Is(err, parl.ErrNilValue) { …
var ErrNilValue = NilValueError(errors.New("end callbacks error"))

func NilValueError(err error) (err2 error) {
	return nilValue{err}
}

type nilValue struct{ error }

func (ec *nilValue) Is(err error) (is bool) {
	_, is = err.(interface{ NilValueError() })
	return
}
func (ec *endCallbacks) NilValueError() {}
