/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package recover

import (
	"errors"
)

// ErrEndCallbacks indicates upon retun from a callback function that
// no more callbacks are desired. It does not indicate an error and is not returned
// as an error by any other function than the callback.
//
// callback invocations may be thread-safe, re-entrant and panic-handling but
// this depends on the callback-invoking implementation used.
//
// Usage:
//
//	if errors.Is(err, parl.ErrEndCallbacks) { …
var ErrEndCallbacks = EndCallbacks(errors.New("end callbacks error"))

func EndCallbacks(err error) (err2 error) {
	return endCallbacks{err}
}

type endCallbacks struct{ error }

func (ec *endCallbacks) Is(err error) (is bool) {
	_, is = err.(interface{ EndCallbacks() })
	return
}
func (ec *endCallbacks) EndCallbacks() {}
