/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// AddError is a function to submit non-fatal errors
type AddError func(err error)

// AddErrorIf is an interface implementing the AddError function
type AddErrorIf interface {
	// AddError is a function to submit non-fatal errors
	AddError(err error)
}
