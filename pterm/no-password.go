/*
© 2021–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pterm

import (
	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/perrors"
)

// [NoPassword] is a value indicating that password input is unavailable
type NoPasswordStruct struct{}

// [NoPassword] is a value indicating that password input is unavailable
var NoPassword = &NoPasswordStruct{}

var _ parl.Password = &NoPasswordStruct{}

// HasPassword indicates password input not available
func (ps *NoPasswordStruct) HasPassword() (hasPassword bool) { return false }

// Password return empty password
func (ps *NoPasswordStruct) Password() (password []byte, err error) {
	err = perrors.NewPF("password input unavailable")
	return
}
