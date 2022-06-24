/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

type DoErr interface {
	DoErr(op func() (err error)) (err error)
}
