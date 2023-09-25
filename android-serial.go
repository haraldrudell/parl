/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

func (a AndroidSerial) String() (s string) {
	return string(a)
}

func (a AndroidSerial) IsValid() (isValid bool) {
	return len(string(a)) > 0
}