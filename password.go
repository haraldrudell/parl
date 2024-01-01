/*
© 2020–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// Password indicates if a password can be read interactively
//   - [pterm.NoPassword] implements password input unavailable
//   - [pterm.NewPassword] implements interactive password input
type Password interface {
	// true if a password can be obtained interactively
	HasPassword() (hasPassword bool)
	// blocking interactive password input
	Password() (password []byte, err error)
}
