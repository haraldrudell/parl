/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// NewAndroidStatus returns Anddroid status of s
//   - AndroidStatus is a single word of ANSII-set characters
func NewAndroidStatus(s string) (status AndroidStatus) { return AndroidStatus(s) }

// IsValid returns whether a conatins a valid Android device status
func (a AndroidStatus) IsValid() (isValid bool) { return len(string(a)) > 0 }

// IsOnline returns whether the Android status is device online, ie. ready for interactions
func (a AndroidStatus) IsOnline() (isOnline bool) { return a == AndroidOnline }

func (a AndroidStatus) String() (s string) { return string(a) }
