/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import (
	"errors"
	"fmt"
)

// Unwrapper unwraps error created by [fmt.Errorf]
//   - Unwrapper provides traversal of an error chain
//   - it is the package-private interface used by [errors.Unwrap]
//   - go1.13 190903
type Unwrapper interface {
	// Unwrap returns the result of calling the Unwrap method on err, if err's
	// type contains an Unwrap method returning error.
	// Otherwise, Unwrap returns nil.
	//
	// Unwrap only calls a method of the form "Unwrap() error".
	// In particular Unwrap does not unwrap errors returned by [Join].
	Unwrap() (err error)
}

// func fmt.Errorf(format string, a ...any) error
var _ = fmt.Errorf

// [errors.Unwrap] is legacy unwrap prior to Join
//   - func errors.Unwrap(err error) error
var _ = errors.Unwrap
