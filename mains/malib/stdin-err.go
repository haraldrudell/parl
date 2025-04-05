/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package malib

// Err is type-name for error,
// enabling error as promoted public field
type Err error

// StdinErr is an error value wrapping an error or panic
//   - consumer can determine whether error or panic
//   - unadulterated original error is available
type StdinErr struct {
	//   - RecoverAny non-nil: the any-typed value from
	//     recover built-in due to panic
	//   - RecoverAny nil: Err is error from [os.Stdin.Read]
	RecoverAny any
	// - Err: error from os.Stdin.Read including [io.EOF]
	// - — never nil
	// - — if RecoverAny is error, it is assigned to Err
	// - — if recoverAny is not error, a place-holder string error
	Err
}
