/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

type NoPasswordStruct struct{}

var NoPassword = &NoPasswordStruct{}

func (ps *NoPasswordStruct) HasPassword() (hasPassword bool) {
	return false
}

func (ps *NoPasswordStruct) Password() (password string) {
	return ""
}
