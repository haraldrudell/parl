/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

func NewAndroidStatus(s string) (status AndroidStatus) {
	return AndroidStatus(s)
}

func (a AndroidStatus) String() (s string) {
	return string(a)
}

func (a AndroidStatus) IsValid() (isValid bool) {
	return len(string(a)) > 0
}

func (a AndroidStatus) IsOnline() (isOnline bool) {
	return a == AndroidOnline
}
