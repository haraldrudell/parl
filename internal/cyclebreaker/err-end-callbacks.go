/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package cyclebreaker

import (
	"errors"
)

// ErrEndCallbacks indicates upon retun from a callback function that
// no more callbacks are desired
//   - ErrEndCallbacks does not indicate an error and should not be returned
//     by any other function than a callback
//   - check for ErrrEndCallback is using the [errors.Is] method with the
//     [parl.ErrEndCallbacks] value
//   - EndCallbacks creates an ErrEndCallbacks value basd on another error
//   - —
//   - an ErrEndCallbacks type implements:
//   - — an Is method returning true for errors implementing a EndCallbacks method
//   - parl.ErrEndCallbacks additionally implements:
//   - — a dummy EndCallbacks method
//
// Usage:
//
//	func callback() (err error) {
//	  return parl.ErrEndCallbacks
//
//	  if errors.Is(err, parl.ErrEndCallbacks) {
//	    err = nil
//	    …
var ErrEndCallbacks = EndCallbacks(errors.New("end callbacks error"))

// EndCallbacks creates a EndCallbacks error based on another error
func EndCallbacks(err error) (err2 error) { return &endCallbacks{err} }

// endCallbacks is the named type of [parl.ErrEndCallbacks] value
type endCallbacks struct{ error }

var _ = errors.Is

// Is indicates that endCallbacks type is EndCallBack
func (e *endCallbacks) Is(err error) (is bool) {
	_, is = err.(interface{ EndCallbacks() })
	return
}

// dummy EndCallbacks method
func (e *endCallbacks) EndCallbacks() {}
