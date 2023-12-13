/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package sqliter

import "errors"

// ErrDsnNotExist can detect DSN not exist errors
//
// Usage:
//
//	if errors.Is(err, sqliter.ErrDsnNotExist) {
var ErrDsnNotExist = &dsnNotExist{error: errors.New("ErrDsnNotExist")}

// MarkDsnNotExist marks error as cased by read-only DSN
func MarkDsnNotExist(err error) (err2 error) {
	return &dsnNotExist{error: err}
}

// endCallbacks is the named type of [parl.ErrEndCallbacks] value
type dsnNotExist struct{ error }

// Is indicates that endCallbacks type is EndCallBack
func (e *dsnNotExist) Is(err error) (is bool) {
	_, is = err.(interface{ DSNNotExist() })
	return
}

// Unwrap of error chain
func (e *dsnNotExist) Unwrap() (err error) { return e.error }

// dummy EndCallbacks method
func (e *dsnNotExist) DSNNotExist() {}
