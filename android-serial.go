/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// NewAndroidSerial returns Android serial for s
//   - typically a string of a dozen or so 8-bit chanacters consisting of
//     lower and upper case a-zA-Z0-9
func NewAndroidSerial(s string) (serial AndroidSerial) { return AndroidSerial(s) }

// IsValid() returns whether a contains a valid Android serial
func (a AndroidSerial) IsValid() (isValid bool) { return len(string(a)) > 0 }

func (a AndroidSerial) String() (s string) { return string(a) }
