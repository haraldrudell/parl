/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package errorglue

import "errors"

// JoinUnwrapper unwraps error created by [errors.Join]
//   - go1.20 230201
type JoinUnwrapper interface {
	// Unwrap returns the list of non-nil errors provided to
	// [errors.Join]
	Unwrap() (err []error)
}

// similar to [Unwrapper]
var _ Unwrapper

// [errors.Join]
//   - func Join(errs ...error) error
//   - joinError type is private to errors package
var _ = errors.Join

// Joined errors work with [errors.Is] and [errors.As]
//   - traversal order: all erorrs in zero-chain, then the innermost
//     join errors of each chain being traversed
var _, _ = errors.Is, errors.As
