/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type AddError func(err error)

type AddErrorIf interface {
	AddError(err error)
}
